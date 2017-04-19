package data

import (
	"strings"
	"fmt"
	"encoding/json"
)

type CardData struct {
	// TODO ?->  ID int `json:"databaseId"`
	Name                    string        `json:"id"`
	Portrait                string        `json:"portrait"`
	Rarity                  string        `json:"rarity"`
	Tier                    int           `json:"tier,string"`
	Type                    string        `json:"type"`
	Units                   []string      `json:"units"`
	UnitCount               int           `json:"numUnits,string"`
	ManaCost                int           `json:"manaCost,string"`
	Cooldown                int           `json:"cooldown,string"`
	AwakenGamesNeeded       int           `json:"awakenGamesNeeded,string"`
	AwakenLeaderGamesNeeded int           `json:"awakenLeaderGamesNeeded,string"`
}

// data map
var cards map[DataId]*CardData

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
	json.Unmarshal(raw, container)

	// enter into system data
	cards = map[DataId]*CardData {}
	for i, card := range container.Cards {
		name := card.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		if err != nil {
			panic(err)
		}

		// insert into table
		cards[id] = &container.Cards[i]
	}
}

// get card by server ID
func GetCard(id DataId) (card *CardData) {
	return cards[id]
}

func GetCardsByTieredRarity(tier int, rarity string) []DataId {
	cardSlice := make([]DataId, 0)

	for id, cardData := range cards {
		if cardData.Tier <= tier && cardData.Rarity == rarity {
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