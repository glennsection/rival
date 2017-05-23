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
)

// store category
type StoreCategory int
const (
	StoreCategoryPremiumCurrency StoreCategory = iota
	StoreCategoryStandardCurrency
	StoreCategoryTomes
	StoreCategoryCards
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
}

// client data
type StoreDataClientAlias StoreData
type StoreDataClient struct {
	Category                string        `json:"category"`
	Currency                string        `json:"currency"`

	*StoreDataClientAlias
}

type CardPurchaseCost struct {
	Rarity 					string 		  `json:"rarity"`
	Cost 					string 		  `json:"cost"`
}

// store item data map
var storeItems map[DataId]*StoreData

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

	// server category
	switch client.Category {
	case "PremiumCurrency":
		storeItem.Category = StoreCategoryPremiumCurrency
	case "Tomes":
		storeItem.Category = StoreCategoryTomes
	case "Cards":
		storeItem.Category = StoreCategoryCards
	default:
		storeItem.Category = StoreCategoryStandardCurrency
	}

	// server currency
	switch client.Currency {
	case "Real":
		storeItem.Currency = CurrencyReal
	default:
		storeItem.Currency = CurrencyPremium
	}

	return nil
}

// data processor
func LoadStore(raw []byte) {
	// parse
	container := &StoreParsed {}
	json.Unmarshal(raw, container)

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
	json.Unmarshal(raw, container)

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

// get card cost by rarity and potential level
func GetCardCost(rarity string, level int) int {
	level -= 1

	if level > len(cardPurchaseCosts[rarity]) {
		return -1
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
