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
	StoreCategorySpecialOffers
)

type AvailabilityType int
const (
	Availability_Permanent AvailabilityType = iota
	Availability_Limited
)

// server data
type StoreItemData struct {
	Name                    string
	DisplayName 			string
	Description 			string

	ItemID                  string
	Category                StoreCategory
	RewardIDs 				[]DataId

	Currency                CurrencyType
	Cost                    float64

	Availability 			AvailabilityType
	League 					int
	IsOneTimeOffer 			bool
	AvailableDate 			int64
	ExpirationDate 			int64
	Duration 				int
}

// client data
type StoreItemDataClientAlias StoreItemData
type StoreItemDataClient struct {
	Name                    string        `json:"id"`
	DisplayName 			string 		  `json:"displayName"`
	Description 			string 		  `json:"description"`

	ItemID                  string        `json:"itemId"`
	Category                string        `json:"category"`
	RewardIDs				string 	  	  `json:"rewardIds"`

	Currency                string        `json:"currency"`
	Cost                    float64       `json:"cost,string"`

	Availability 			string 		  `json:"availability"`
	League 					string 		  `json:"league"`
	IsOneTimeOffer 			bool 		  `json:"isOneTimeOffer,string"`
	AvailableDate 			string 		  `json:"availableDate"`
	ExpirationDate 			string 		  `json:"expirationDate"`
	Duration 				string 		  `json:"duration"`
}

type CardPurchaseCost struct {
	Rarity 					string 		  `json:"rarity"`
	Cost 					string 		  `json:"cost"`
}

// store item data map
var storeItems map[DataId]*StoreItemData

// card purchasing data map
var cardPurchaseCosts map[string][]int

// implement Data interface
func (data *StoreItemData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type StoreParsed struct {
	Store []StoreItemData
}

type CardPurchaseCostsParsed struct {
	CardPurchaseCosts []CardPurchaseCost
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
	storeItemData.DisplayName = client.DisplayName
	storeItemData.Description = client.Description
	storeItemData.ItemID = client.ItemID
	storeItemData.Cost = client.Cost
	storeItemData.IsOneTimeOffer = client.IsOneTimeOffer

	// server reward ids
	storeItemData.RewardIDs = make([]DataId, 0)

	clientRewards := util.StringToStringArray(client.RewardIDs)
	for _,id := range clientRewards {
		storeItemData.RewardIDs = append(storeItemData.RewardIDs, ToDataId(id))
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
	
	// server availability
	if storeItemData.Availability, err = StringToAvailabilityType(client.Availability); err != nil {
		panic(err)
	}

	// server League
	if num, err := strconv.ParseInt(client.League, 10, 32); err == nil {
		storeItemData.League = int(num)
	} else {
		storeItemData.League = 0
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
		storeItemData.Duration = 0
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
	for i, storeItem := range container.Store {
		name := storeItem.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		util.Must(err)

		// insert into table
		storeItems[id] = &container.Store[i]
	}
}

func LoadCardPurchaseCosts(raw []byte) {
	//parse
	container := &CardPurchaseCostsParsed {}
	util.Must(json.Unmarshal(raw, container))

	//enter into system data
	cardPurchaseCosts = map[string][]int{}
	for _, data := range container.CardPurchaseCosts {
		cardPurchaseCosts[data.Rarity] = util.StringToIntArray(data.Cost)
	}
}

// get store item by server ID
func GetStoreItemData(id DataId) (store *StoreItemData) {
	return storeItems[id]
}

func GetStoreItemDataCollection() (map[DataId]*StoreItemData) {
	return storeItems
}

// cards can't be purchased past a certain level specific to each rarity. this function
// determins if a given card of a provided level is eligible for purchase
func CanPurchaseCard(rarity string, level int) bool {
	level -= 1

	return level <= len(cardPurchaseCosts[rarity])
}

// get card cost by rarity and potential level
func GetCardCost(rarity string, level int) int {
	level -= 1

	if level >= len(cardPurchaseCosts[rarity]) {
		level = len(cardPurchaseCosts[rarity]) - 1
	}

	return cardPurchaseCosts[rarity][level]
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
	case StoreCategorySpecialOffers:
		return "SpecialOffers", nil
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
	case "SpecialOffers":
		return StoreCategorySpecialOffers, nil
	}

	return StoreCategorySpecialOffers, errors.New(fmt.Sprintf("Cannot convert %s to StoreCategory", val))
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

func AvailabilityTypeToString(val AvailabilityType) (string, error) {
	switch val {
	case Availability_Limited:
		return "Limited", nil
	case Availability_Permanent:
		return "Permanent", nil
	}

	return "", errors.New("Invalid value passed as AvailabilityType")
}

func StringToAvailabilityType(val string) (AvailabilityType, error) {
	switch val {
	case "Limited":
		return Availability_Limited, nil
	case "Permanent":
		return Availability_Permanent, nil
	}

	return Availability_Permanent, errors.New(fmt.Sprintf("Cannot convert %s to AvailabilityType", val))
}