package data

import (
	"encoding/json"
)

type TomeData struct {
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
var tomes map[DataId]TomeData

// implement Data interface
func (data TomeData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type TomesParsed struct {
	Tomes []TomeData
}

// data processor
func LoadTomes(raw []byte) {
	// parse
	container := &TomesParsed {}
	json.Unmarshal(raw, container)

	// enter into system data
	tomes = map[DataId]TomeData {}
	for _, tome := range container.Tomes {
		name := tome.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		if err != nil {
			panic(err)
		}

		// insert into table
		tomes[id] = tome
	}
}

// get tome by server ID
func GetTome(id DataId) (tome TomeData) {
	return tomes[id]
}