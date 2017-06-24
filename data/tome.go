package data

import (
	"fmt"
	"encoding/json"
	"strings"

	"bloodtales/util"
)

// server data
type TomeData struct {
	Name                    string        `json:"id"`
	Image                   string        `json:"icon"`
	Rarity                  string        `json:"rarity"`
	Chance 					float64		  `json:"chance,string"`
	TimeToUnlock			int 		  `json:"timeToUnlock,string"`
	GemsToUnlock			int 		  `json:"gemsToUnlock,string"`
	RewardID 				DataId 		  
}

// client data
type TomeDataClientAlias TomeData
type TomeDataClient struct {
	RewardID 				string 		  `json:"rewardId"`

	*TomeDataClientAlias
}

type TomeOrderClient struct {
	ID 						string 		  `json:"id"`
}

// data map
var tomes map[DataId]*TomeData

var tomeOrder []DataId

// implement Data interface
func (data *TomeData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
type TomesParsed struct {
	Tomes []TomeData
}

type TomeOrderParsed struct {
	TomeOrder []TomeOrderClient
}

// custom unmarshalling
func (tome *TomeData) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &TomeDataClient {
		TomeDataClientAlias: (*TomeDataClientAlias)(tome),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server rewards
	tome.RewardID = ToDataId(client.RewardID)

	return nil
}

// data processor
func LoadTomes(raw []byte) {
	// parse
	container := &TomesParsed {}
	util.Must(json.Unmarshal(raw, container))

	// enter into system data
	tomes = map[DataId]*TomeData {}
	for i, tome := range container.Tomes {
		name := tome.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		util.Must(err)

		// insert into table
		tomes[id] = &container.Tomes[i]
	}
}

func LoadTomeOrder(raw []byte) {
	// parse
	container := &TomeOrderParsed {}
	util.Must(json.Unmarshal(raw, container))

	// enter into system data
	tomeOrder = make([]DataId, len(container.TomeOrder))
	for i, tome := range container.TomeOrder {
		//convert to data id
		id := ToDataId(tome.ID)

		// set val in slice
		tomeOrder[i] = id
	}
}

// get tome by server ID
func GetTome(id DataId) (tome *TomeData) {
	return tomes[id]
}

// get the next tome the player has earned for winning a match
func GetNextVictoryTomeID(winCount int) DataId {
	fmt.Println(fmt.Sprintf("WIN COUNT: %d, TOME ID: %s", winCount, ToDataName(tomeOrder[(winCount - 1) % len(tomeOrder)])))

	return tomeOrder[(winCount - 1) % len(tomeOrder)]
}

func (tome *TomeData) GetImageSrc() string {
	src := tome.Image
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/tomes/%v.png", src)
}