package data

import (
	"encoding/json"
)

type CardData struct {
	// TODO ?->  ID int `json:"databaseId"`
	Name                    string        `json:"id"`
	Rarity                  string        `json:"rarity"`
	Tier                    int           `json:"tier"`
	Type                    string        `json:"type"`
	Units                   []string      `json:"units"`
	UnitCount               int           `json:"numUnits"`
	ManaCost                int           `json:"manaCost"`
	Cooldown                int           `json:"cooldown"`
	AwakenGamesNeeded       int           `json:"awakenGamesNeeded"`
	AwakenLeaderGamesNeeded int           `json:"awakenLeaderGamesNeeded"`
}

// data map
var cards map[DataId]CardData

// implement Data interface
func (data CardData) GetDataName() string {
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
	cards = map[DataId]CardData {}
	for _, card := range container.Cards {
		name := card.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		if err != nil {
			panic(err)
		}

		// insert into table
		cards[id] = card
	}
}

// get card by server ID
func GetCard(id DataId) (card CardData) {
	return cards[id]
}