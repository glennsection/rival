package data

import (
	"fmt"
	"time"
	"errors"
	"strconv"
	"encoding/json"

	"bloodtales/util"
)

type CurrencyType int
const (
	CurrencyReal CurrencyType = iota
	CurrencyPremium
	CurrencyStandard
)

type StoreCategory int
const (
	StoreCategoryPremiumCurrency StoreCategory = iota
	StoreCategoryStandardCurrency
	StoreCategoryTomes
	StoreCategoryCards
	StoreCategoryOneTimeOffers
	StoreCategoryPeriodicOffers
)

type OfferPriority int
const (
	OfferPriority_Highest OfferPriority = iota
	OfferPriority_High 
	OfferPriority_Medium 
	OfferPriority_Low 
	OfferPriority_Lowest
)

// server data
type StoreItemData struct {
	Name                    string

	ItemID                  string
	Category                StoreCategory
	RewardIDs 				[]DataId
	RewardLevel 			int

	Currency                CurrencyType
	Cost                    float64

	Priority 				OfferPriority
	Leagues 				map[League]interface{} //should be used as a hashset, interface value will always be nil
	LevelRequirement 		int
	AvailableDate 			int64
	ExpirationDate 			int64
	Duration 				int
	Cooldown 				int
}

// client data
type StoreItemDataClientAlias StoreItemData
type StoreItemDataClient struct {
	Name                    string        	`json:"id"`

	ItemID                  string        	`json:"itemId"`
	Category                string        	`json:"category"`
	RewardIDs				string 	  	  	`json:"rewardIds"`
	RewardLevel 			string 			`json:"rewardLevel"`

	Currency                string        	`json:"currency"`
	Cost                    float64       	`json:"cost,string"`

	Priority 				string 		  	`json:"priority"`
	Leagues 				string 		  	`json:"leagues"`
	LevelRequirement 		string 		  	`json:"levelRequirement"`
	AvailableDate 			string 		  	`json:"availableDate"`
	ExpirationDate 			string 		  	`json:"expirationDate"`
	Duration 				string 		  	`json:"duration"`
	Cooldown 				string 			`json:"cooldown"`
}

type PeriodicOfferClient struct {
	ID 						string 			`json:"id"`
}

// store item data map
var storeItems map[DataId]*StoreItemData

// special offer data map
var oneTimeOffers map[DataId]*StoreItemData

// periodic offer data map
var periodicOffers map[DataId]*StoreItemData

// periodic offer table
var periodicOfferTable []DataId

// implement Data interface
func (data *StoreItemData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type StoreParsed struct {
	Store []StoreItemData
}

type PeriodicOfferTableParsed struct {
	PeriodicOfferTable []PeriodicOfferClient
}

// custom unmarshalling
func (storeItemData *StoreItemData) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &StoreItemDataClient {} //alias doesn't work for some reason

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	storeItemData.Name = client.Name
	storeItemData.ItemID = client.ItemID
	storeItemData.Cost = client.Cost

	// server reward ids
	storeItemData.RewardIDs = make([]DataId, 0)

	clientRewards := util.StringToStringArray(client.RewardIDs)
	for _, id := range clientRewards {
		storeItemData.RewardIDs = append(storeItemData.RewardIDs, ToDataId(id))
	}

	// server reward level
	if num, err := strconv.ParseInt(client.RewardLevel, 10, 64); err == nil {
		storeItemData.RewardLevel = int(num)
	} else {
		storeItemData.RewardLevel = 0
	}

	var err error
	err = nil

	// server category
	if storeItemData.Category, err = StringToStoreCategory(client.Category); err != nil {
		panic(err)
	} 

	// server currency
	if storeItemData.Currency, err = StringToCurrencyType(client.Currency); err != nil {
		panic(err)
	}

	// server priority
	if storeItemData.Category == StoreCategoryOneTimeOffers || storeItemData.Category == StoreCategoryPeriodicOffers {
		if storeItemData.Priority, err = StringToOfferPriority(client.Priority); err != nil {
			panic(err)
		}
	}

	// server Leagues
	storeItemData.Leagues = map[League]interface{}{}

	clientLeagues := util.StringToStringArray(client.Leagues)
	for _, league := range clientLeagues {
		storeItemData.Leagues[GetLeagueByID(league)] = nil
	}

	// server level requirement
	if num, err := strconv.ParseInt(client.LevelRequirement, 10, 64); err == nil {
		storeItemData.LevelRequirement = int(num)
	} else {
		storeItemData.LevelRequirement = 0
	}

	// server AvailableDate
	if num, err := strconv.ParseInt(client.AvailableDate, 10, 64); err == nil {
		storeItemData.AvailableDate = num
	} else {
		if date, err := time.Parse("2006-01-02", client.AvailableDate); err == nil {
			storeItemData.AvailableDate = util.TimeToTicks(date)
		} else {
			storeItemData.AvailableDate = 0
		}
	}

	// server ExpirationDate
	if num, err := strconv.ParseInt(client.ExpirationDate, 10, 64); err == nil {
		storeItemData.ExpirationDate = num
	} else {
		if date, err := time.Parse("2006-01-02", client.ExpirationDate); err == nil {
			storeItemData.ExpirationDate = util.TimeToTicks(date)
		} else {
			storeItemData.ExpirationDate = 0
		}
	}

	// server Duration
	if num, err := strconv.ParseInt(client.Duration, 10, 32); err == nil {
		storeItemData.Duration = int(num)
	} else {
		storeItemData.Duration = 1
	}

	// server cooldown
	if num, err := strconv.ParseInt(client.Cooldown, 10, 32); err == nil {
		storeItemData.Cooldown = int(num)
	} else {
		storeItemData.Cooldown = -1
	}

	return nil
}

// data processor
func LoadStore(raw []byte) {
	// parse
	container := &StoreParsed {}
	util.Must(json.Unmarshal(raw, container))

	// enter into system data
	storeItems = map[DataId]*StoreItemData {}
	oneTimeOffers = map[DataId]*StoreItemData {}
	periodicOffers = map[DataId]*StoreItemData {}

	for i, storeItem := range container.Store {
		name := storeItem.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		util.Must(err)

		// insert into appropriate table
		switch storeItem.Category {
		case StoreCategoryPeriodicOffers:
			periodicOffers[id] = &container.Store[i]
			break;
		case StoreCategoryOneTimeOffers:
			oneTimeOffers[id] = &container.Store[i]
			break;
		default:
			storeItems[id] = &container.Store[i]
		}
	}
}

func LoadPeriodicOfferTable(raw []byte) {
	// parse
	container := &PeriodicOfferTableParsed {}
	util.Must(json.Unmarshal(raw, container))

	//enter into system data
	periodicOfferTable = make([]DataId, len(container.PeriodicOfferTable))

	for i, offer := range container.PeriodicOfferTable {
		//convert to data id
		id := ToDataId(offer.ID)

		// set val in slice
		periodicOfferTable[i] = id
	}
}

// get store item by server ID
func GetStoreItemData(id DataId) (store *StoreItemData) {
	if storeItem, contains := storeItems[id]; contains {
		return storeItem
	}

	if specialOffer, contains := oneTimeOffers[id]; contains {
		return specialOffer
	} 
	
	if periodicOffer, contains := periodicOffers[id]; contains {
		return periodicOffer
	}

	return nil
}

func GetRegularStoreCollection() (map[DataId]*StoreItemData) {
	return storeItems
}

func GetOneTimeOfferCollection() (map[DataId]*StoreItemData) {
	return oneTimeOffers
}

func GetPeriodicOfferCollection() (map[DataId]*StoreItemData) {
	return periodicOffers
}

func GetPeriodicOfferTable() []DataId {
	return periodicOfferTable
}

func StoreCategoryToString(val StoreCategory) (string, error) {
	switch val {
	case StoreCategoryPremiumCurrency:
		return "PremiumCurrency", nil
	case StoreCategoryTomes:
		return "Tomes", nil
	case StoreCategoryCards:
		return "Cards", nil
	case StoreCategoryStandardCurrency:
		return "StandardCurrency", nil
	case StoreCategoryOneTimeOffers:
		return "OneTimeOffers", nil
	case StoreCategoryPeriodicOffers:
		return "PeriodicOffers", nil
	}
	
	return "", errors.New("Invalid value passed as StoreCategory")
}

func StringToStoreCategory(val string) (StoreCategory, error) {
	switch val {
	case "PremiumCurrency":
		return StoreCategoryPremiumCurrency, nil
	case "Tomes":
		return StoreCategoryTomes, nil
	case "Cards":
		return StoreCategoryCards, nil
	case "StandardCurrency":
		return StoreCategoryStandardCurrency, nil
	case "OneTimeOffers":
		return StoreCategoryOneTimeOffers, nil
	case "PeriodicOffers":
		return StoreCategoryPeriodicOffers, nil
	}

	return StoreCategoryPeriodicOffers, errors.New(fmt.Sprintf("Cannot convert %s to StoreCategory", val))
}

func CurrencyTypeToString(val CurrencyType) (string, error) {
	switch val {
	case CurrencyReal:
		return "Real", nil
	case CurrencyPremium:
		return "Premium", nil
	case CurrencyStandard:
		return "Standard", nil
	}

	return "", errors.New("Invalid value passed as CurrencyType")
}

func StringToCurrencyType(val string) (CurrencyType, error) {
	switch val {
	case "Real":
		return CurrencyReal, nil
	case "Premium":
		return CurrencyPremium, nil
	case "Standard":
		return CurrencyStandard, nil
	}

	return CurrencyStandard, errors.New(fmt.Sprintf("Cannot convert %s to CurrencyType", val))
}

func OfferPriorityToString(val OfferPriority) (string, error) {
	switch val {
	case OfferPriority_Highest:
		return "Highest", nil
	case OfferPriority_High:
		return "High", nil
	case OfferPriority_Medium:
		return "Medium", nil
	case OfferPriority_Low:
		return "Low", nil
	case OfferPriority_Lowest:
		return "Lowest", nil
	}

	return "", errors.New("Invalid value passed as OfferPriority")
}

func StringToOfferPriority(val string) (OfferPriority, error) {
	switch val {
	case "Highest":
		return OfferPriority_Highest, nil
	case "High":
		return OfferPriority_High, nil
	case "Medium":
		return OfferPriority_Medium, nil
	case "Low":
		return OfferPriority_Low, nil
	case "Lowest":
		return OfferPriority_Lowest, nil
	}

	return OfferPriority_Lowest, errors.New(fmt.Sprintf("Cannot convert %s to OfferPriority", val))
}