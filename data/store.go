package data

import (
	// "fmt"
	// "strconv"
	// "strings"
	// "encoding/json"
)

// currency type
type CurrencyType int
const (
	CurrencyReal CurrencyType = iota
	CurrencyPremium
)
   // "id": "STORE_PREMIUM_1",
   //  "category": "PremiumCurrency",
   //  "displayName": "Pile",
   //  "spritePath": "UI/Store/Item Icons/store_blood_pack_1",
   //  "itemId": "",
   //  "quantity": "30",
   //  "currency": "Real",
   //  "cost": "2.99"

type StoreData struct {
	Name                    string        `json:"id"`
	Image                   string        `json:"category"`
	Rarity                  string        `json:"itemId"`
	Chance                  int           `json:"quantity,string"`
	Currency                int           `json:"currency,string"`
	Cost                    int           `json:"cost,string"`
}

/*
// data map
var tomes map[DataId]*TomeData

// implement Data interface
func (data *TomeData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type TomesParsed struct {
	Tomes []RawTomeData
}

func (rawTomeData *RawTomeData) ToTomeData() (tomeData *TomeData) {
	tomeData = &TomeData{
		Name: rawTomeData.Name,
		Image: rawTomeData.Image,
		Rarity: rawTomeData.Rarity,
		Chance: rawTomeData.Chance,
		TimeToUnlock: rawTomeData.TimeToUnlock,
		GemsToUnlock: rawTomeData.GemsToUnlock,
		MinPremiumReward: rawTomeData.MinPremiumReward,
		MaxPremiumReward: rawTomeData.MaxPremiumReward,
		MinStandardReward: rawTomeData.MinStandardReward,
		MaxStandardReward: rawTomeData.MaxStandardReward,
		GuaranteedRarities: []int {0, 0, 0, 0},
		CardsRewarded: []int {0, 0, 0, 0},
	}

	// convert string formatted array to []int
	guaranteedRarities := strings.FieldsFunc(rawTomeData.GuaranteedRarities, func (r rune) bool {
		return r == '[' || r == ',' || r == ']'
	})
	for i, num := range guaranteedRarities {
		tomeData.GuaranteedRarities[i], _ = strconv.Atoi(num)
	}

	// convert string formatted array to []int
	cardsRewarded := strings.FieldsFunc(rawTomeData.CardsRewarded, func (r rune) bool {
		return r == '[' || r == ',' || r == ']'
	})
	for i, num := range cardsRewarded {
		tomeData.CardsRewarded[i], _ = strconv.Atoi(num)
	}

	return
}

// data processor
func LoadTomes(raw []byte) {
	// parse
	container := &TomesParsed {}
	json.Unmarshal(raw, container)

	// enter into system data
	tomes = map[DataId]*TomeData {}
	for _, tome := range container.Tomes {
		tomeData := tome.ToTomeData()
		name := tomeData.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		if err != nil {
			panic(err)
		}

		// insert into table
		tomes[id] = tomeData
	}
}

// get tome by server ID
func GetTome(id DataId) (tome *TomeData) {
	return tomes[id]
}

func GetTomeIdsSorted(compare func(*TomeData, *TomeData) bool) (tomeIds []DataId){
	tomeIds = make([]DataId, 0)

	for id, tomeData := range tomes {
		if len(tomeIds) == 0 {
			tomeIds = append(tomeIds, id)
		} else {
			for i, dataId := range tomeIds {

				if compare(tomeData, tomes[dataId]) {
					tomeIds = append(tomeIds, id)
					copy(tomeIds[i+1:], tomeIds[i:])
					tomeIds[i] = id
					break
				}

				if i == (len(tomeIds) - 1) {
					tomeIds = append(tomeIds, id)
				} 
			}
		}
	}

	return 
}

func (tome *TomeData) GetImageSrc() string {
	src := tome.Image
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/tomes/%v.png", src)
}
*/