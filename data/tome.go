package data

import (
	"fmt"
	"strings"
	"encoding/json"
)

type TomeData struct {
	Name                    string        `json:"id"`
	Image                   string        `json:"icon"`
	Rarity                  string        `json:"rarity"`
	TimeToUnlock			int 		  `json:"timeToUnlock,string"`
	GemsToUnlock			int 		  `json:"gemsToUnlock,string"`
	MinPremiumReward		int 		  `json:"minGemReward,string"`
	MaxPremiumReward		int 		  `json:"maxGemReward,string"`
	MinStandardReward		int 		  `json:"minGoldReward,string"`
	MaxStandardReward		int 		  `json:"maxGoldReward,string"`
	GuaranteedRarities		[]int		  `json:"guaranteedRarities,string"`
	CardsRewarded			[]int		  `json:"cardsRewarded,string"`
}

// data map
var tomes map[DataId]*TomeData

// implement Data interface
func (data *TomeData) GetDataName() string {
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
	tomes = map[DataId]*TomeData {}
	for i, tome := range container.Tomes {
		name := tome.GetDataName()

		// map name to ID
		id, err := mapDataName(name)
		if err != nil {
			panic(err)
		}

		// insert into table
		tomes[id] = &container.Tomes[i]
	}
}

// get tome by server ID
func GetTome(id DataId) (tome *TomeData) {
	return tomes[id]
}

func (tome *TomeData) GetImageSrc() string {
	src := tome.Image
	idx := strings.LastIndex(src, "/")
	if idx >= 0 {
		src = src[idx + 1:]
	}
	return fmt.Sprintf("/static/img/tomes/%v.png", src)
}