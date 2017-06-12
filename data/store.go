package data

import (
	"fmt"
	"strings"
	"encoding/json"

	"bloodtales/util"
)

// currency type
type CurrencyType int
const (
	CurrencyReal CurrencyType = iota
	CurrencyPremium
	CurrencyStandard
)

// store category
type StoreCategory int
const (
	StoreCategoryPremiumCurrency StoreCategory = iota
	StoreCategoryStandardCurrency
	StoreCategoryTomes
	StoreCategoryCards
	StoreCategorySpecialOffers
)

// server data
type StoreData struct {
	Name                    string        `json:"id"`
	Category                StoreCategory `json:"category"`
	DisplayName 			string 		  `json:"displayName"`
	Image                   string        `json:"spritePath"`
	ItemID                  string        `json:"itemId"`
	Quantity                int           `json:"quantity,string"`
	Currency                CurrencyType  `json:"currency"`
	Cost                    float64       `json:"cost,string"`
	RewardID 				DataId
}

// client data
type StoreDataClientAlias StoreData
type StoreDataClient struct {
	Name                    string        `json:"id"`
	Category                string        `json:"category"`
	DisplayName 			string 		  `json:"displayName"`
	Image                   string        `json:"spritePath"`
	ItemID                  string        `json:"itemId"`
	Quantity                int           `json:"quantity,string"`
	Currency                string        `json:"currency"`
	Cost                    float64       `json:"cost,string"`
	RewardID				string 		  `json:"rewardId"`

	*StoreDataClientAlias
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
	client := &StoreDataClient {
		StoreDataClientAlias: (*StoreDataClientAlias)(storeItem),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	//alias doesn't work for some reason
	storeItem.Name = client.Name
	storeItem.DisplayName = client.DisplayName
	storeItem.Image = client.Image
	storeItem.ItemID = client.ItemID
	storeItem.Quantity = client.Quantity
	storeItem.Cost = client.Cost
	storeItem.RewardID = ToDataId(client.RewardID)

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

	return nil
}

// custom marshalling
func (storeItem *StoreData) MarshalJSON() ([]byte, error) {
	client := &StoreDataClient {
		Name: storeItem.Name,
		DisplayName: storeItem.DisplayName,
		Image: storeItem.Image,
		ItemID: storeItem.ItemID,
		Quantity: storeItem.Quantity,
		Cost: storeItem.Cost,
		RewardID: ToDataName(storeItem.RewardID),
	}

	//client category
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

	for _, value := range storeItems {
		items = append(items, *value) 
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

	if level > len(cardPurchaseCosts[rarity]) {
		level = len(cardPurchaseCosts[rarity]) - 1
	}

	return cardPurchaseCosts[rarity][level]
}

func (store *StoreData) GetImageSrc() string {
	src := store.Image
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/stores/%v.png", src) // FIXME
}
