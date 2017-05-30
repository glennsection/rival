package data

import (
	"strings"
	"fmt"
	"encoding/json"

	"bloodtales/util"
)

type CardData struct {
	// TODO ?->  ID int `json:"databaseId"`
	Name                    string        `json:"id"`
	Portrait                string        `json:"portrait"`
	Rarity                  string        `json:"rarity"`
	Tier                    int           `json:"tier,string"`
	Type                    string        `json:"type"`
	//Units                   []string      `json:"units"`
	UnitCount               int           `json:"numUnits,string"`
	ManaCost                int           `json:"manaCost,string"`
	Cooldown                int           `json:"cooldown,string"`
	AwakenGamesNeeded       int           `json:"awakenGamesNeeded,string"`
	AwakenLeaderGamesNeeded int           `json:"awakenLeaderGamesNeeded,string"`
}

type CardProgressionData struct {
	Level 					int 		  `json:"level,string"`
	CardsNeeded 			int 		  `json:"cardsNeeded,string"`
	Cost 					int 		  `json:"cost,string"`
	XP 						int 		  `json:"xp,string"`
}

// data map
var cards map[DataId]*CardData

// card progression
var cardLeveling map[string][]CardProgressionData

// implement Data interface
func (data *CardData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type CardsParsed struct {
	Cards []CardData
}

// data processor
func LoadCards(raw []byte) {
	// parse
	container := &CardsParsed {}
	util.Must(json.Unmarshal(raw, container))

	// enter into system data
	cards = map[DataId]*CardData {}
	for i, card := range container.Cards {
		name := card.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		util.Must(err)

		// insert into table
		cards[id] = &container.Cards[i]
	}
}

func LoadCommonCardProgression(raw []byte) {
	LoadCardProgression("COMMON", raw)
}

func LoadRareCardProgression(raw []byte) {
	LoadCardProgression("RARE", raw)
}

func LoadEpicCardProgression(raw []byte) {
	LoadCardProgression("EPIC", raw)
}

func LoadLegendaryCardProgression(raw []byte) {
	LoadCardProgression("LEGENDARY", raw)
}

func LoadCardProgression(rarity string, raw []byte) { 
	if cardLeveling == nil {
		cardLeveling = map[string][]CardProgressionData {}
	}
	
	// parse, NOTE: because our key names are differ between files, we can't use a structure like CardsParsed. instead, this map works just as well
	var container map[string][]CardProgressionData
	util.Must(json.Unmarshal(raw, &container))

	// insert into table
	for _, dataArray := range container {
		cardLeveling[rarity] = dataArray 
	}
}

// get card by server ID
func GetCard(id DataId) (card *CardData) {
	return cards[id]
}

func GetCards(condition func(*CardData) bool) []DataId {
	cardSlice := make([]DataId, 0)

	for id, cardData := range cards {
		if condition(cardData) {
			cardSlice = append(cardSlice, id)
		}
	}

	return cardSlice
}

func (data *CardData) GetPortraitSrc() string {
	src := data.Portrait
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/portraits/%v.png", src)
}

func GetCardProgressionData(rarity string, level int) CardProgressionData {
	return cardLeveling[rarity][level]
}