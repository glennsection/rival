package data

import (
	"time"
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
type StoreData struct {
	Name                    string
	DisplayName 			string
	Description 			string
	SpritePaths             string // no need to convert this to an array since we have no use for it on the server
	SpriteTexts 			string // see above

	ItemID                  string
	Category                StoreCategory
	RewardIDs 				[]DataId

	Currency                CurrencyType
	Cost                    float64

	Availability 			AvailabilityType
	IsOneTimeOffer 			bool
	AvailableDate 			int64
	ExpirationDate 			int64
}

// client data
type StoreDataClientAlias StoreData
type StoreDataClient struct {
	Name                    string        `json:"id"`
	DisplayName 			string 		  `json:"displayName"`
	Description 			string 		  `json:"description"`
	SpritePaths             string        `json:"spritePaths"`
	SpriteTexts 			string 		  `json:"spriteTexts`

	ItemID                  string        `json:"itemId"`
	Category                string        `json:"category"`
	RewardIDs				string 	  	  `json:"rewardIds"`

	Currency                string        `json:"currency"`
	Cost                    float64       `json:"cost,string"`

	Availability 			string 		  `json:"availability"`
	IsOneTimeOffer 			bool 		  `json:"isOneTimeOffer,string"`
	AvailableDate 			string 		  `json:"availableDate"`
	ExpirationDate 			string 		  `json:"expirationDate"`
}

type CardPurchaseCost struct {
	Rarity 					string 		  `json:"rarity"`
	Cost 					string 		  `json:"cost"`
}

// store item data map
var storeItems map[DataId]*StoreData
//var specialOffers map[DataId]*SpecialOffer

// card purchasing data map
var cardPurchaseCosts map[string][]int

// implement Data interface
func (data *StoreData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type StoreParsed struct {
	Store []StoreData
}

type CardPurchaseCostsParsed struct {
	CardPurchaseCosts []CardPurchaseCost
}

// custom unmarshalling
func (storeItem *StoreData) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &StoreDataClient {} //alias doesn't work for some reason

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	storeItem.Name = client.Name
	storeItem.DisplayName = client.DisplayName
	storeItem.Description = client.Description
	storeItem.SpritePaths = client.SpritePaths
	storeItem.SpriteTexts = client.SpriteTexts
	storeItem.ItemID = client.ItemID
	storeItem.Cost = client.Cost
	storeItem.IsOneTimeOffer = client.IsOneTimeOffer

	// server reward ids
	storeItem.RewardIDs = make([]DataId, 0)

	clientRewards := util.StringToStringArray(client.RewardIDs)
	for _,id := range clientRewards {
		storeItem.RewardIDs = append(storeItem.RewardIDs, ToDataId(id))
	}


	// server category
	switch client.Category {
	case "PremiumCurrency":
		storeItem.Category = StoreCategoryPremiumCurrency
	case "Tomes":
		storeItem.Category = StoreCategoryTomes
	case "Cards":
		storeItem.Category = StoreCategoryCards
	case "StandardCurrency":
		storeItem.Category = StoreCategoryStandardCurrency
	default:
		storeItem.Category = StoreCategorySpecialOffers
	}

	// server currency
	switch client.Currency {
	case "Real":
		storeItem.Currency = CurrencyReal
	case "Premium":
		storeItem.Currency = CurrencyPremium
	default:
		storeItem.Currency = CurrencyStandard
	}

	// server availability
	switch client.Availability {
	case "Limited":
		storeItem.Availability = Availability_Limited
	default:
		storeItem.Availability = Availability_Permanent
	}

	if num, err := strconv.ParseInt(client.AvailableDate, 10, 64); err == nil {
		storeItem.AvailableDate = num
	} else {
		if date, err := time.Parse("2006-01-02", client.AvailableDate); err == nil {
			storeItem.AvailableDate = util.TimeToTicks(date)
		}
	}

	if num, err := strconv.ParseInt(client.ExpirationDate, 10, 64); err == nil {
		storeItem.ExpirationDate = num
	} else {
		if date, err := time.Parse("2006-01-02", client.ExpirationDate); err == nil {
			storeItem.ExpirationDate = util.TimeToTicks(date)
		}
	}

	return nil
}

// custom marshalling
func (storeItem *StoreData) MarshalJSON() ([]byte, error) {
	client := &StoreDataClient {
		Name: storeItem.Name,
		DisplayName: storeItem.DisplayName,
		Description: storeItem.Description,
		SpritePaths: storeItem.SpritePaths,
		SpriteTexts: storeItem.SpriteTexts,
		ItemID: storeItem.ItemID,
		Cost: storeItem.Cost,
		IsOneTimeOffer: storeItem.IsOneTimeOffer,
	}

	// client reward ids
	clientRewards := make([]string, 0)
	for _,id := range storeItem.RewardIDs {
		clientRewards = append(clientRewards, ToDataName(id))
	}
	client.RewardIDs = util.StringArrayToString(clientRewards)

	// client category
	switch storeItem.Category {
	case StoreCategoryPremiumCurrency:
		client.Category = "PremiumCurrency"
	case StoreCategoryTomes:
		client.Category = "Tomes"
	case StoreCategoryCards:
		client.Category = "Cards"
	case StoreCategoryStandardCurrency:
		client.Category = "StandardCurrency"
	default:
		client.Category = "SpecialOffers"
	}

	// client currency
	switch storeItem.Currency {
	case CurrencyReal:
		client.Currency = "Real"
	case CurrencyPremium:
		client.Currency = "Premium"
	default:
		client.Currency = "Standard"
	}

	// client availability
	switch storeItem.Availability {
	case Availability_Limited:
		client.Availability = "Limited"
	default:
		client.Availability = "Permanent"
	}

	if storeItem.AvailableDate > 0 {
		client.AvailableDate = strconv.FormatInt(storeItem.AvailableDate - util.TimeToTicks(time.Now().UTC()), 10)
	}

	if storeItem.ExpirationDate > 0 {
		client.ExpirationDate = strconv.FormatInt(storeItem.ExpirationDate - util.TimeToTicks(time.Now().UTC()), 10)
	}

	return json.Marshal(client)
}

// data processor
func LoadStore(raw []byte) {
	// parse
	container := &StoreParsed {}
	util.Must(json.Unmarshal(raw, container))

	// enter into system data
	storeItems = map[DataId]*StoreData {}
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
func GetStoreItem(id DataId) (store *StoreData) {
	return storeItems[id]
}

func GetStoreItems() []StoreData {
	items := make([]StoreData, 0)
	currentTime := util.TimeToTicks(time.Now().UTC())

	for _, value := range storeItems {
		if value.Availability == Availability_Permanent || value.AvailableDate < currentTime && currentTime < value.ExpirationDate {
			items = append(items, *value)
		}
	}

	return items
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
