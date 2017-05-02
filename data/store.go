package data

import (
	"fmt"
	"strings"
	"encoding/json"
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
	Image                   string        `json:"spritePath"`
	Category                StoreCategory `json:"category"`
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

// data map
var storeItems map[DataId]*StoreData

// implement Data interface
func (data *StoreData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type StoreParsed struct {
	Store []StoreData
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
	default:
		storeItem.Category = StoreCategoryStandardCurrency
	case "Tomes":
		storeItem.Category = StoreCategoryTomes
	case "Cards":
		storeItem.Category = StoreCategoryCards
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
		if err != nil {
			panic(err)
		}

		// insert into table
		storeItems[id] = &container.Store[i]
	}
}

// get store item by server ID
func GetStoreItem(id DataId) (store *StoreData) {
	return storeItems[id]
}

func (store *StoreData) GetImageSrc() string {
	src := store.Image
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/stores/%v.png", src) // FIXME
}
