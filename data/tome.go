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
	MinPremiumReward		int 		  `json:"minGemReward,string"`
	MaxPremiumReward		int 		  `json:"maxGemReward,string"`
	MinStandardReward		int 		  `json:"minGoldReward,string"`
	MaxStandardReward		int  		  `json:"maxGoldReward,string"`
	GuaranteedRarities		[]int		  `json:"guaranteedRarities"`
	CardsRewarded			[]int		  `json:"cardsRewarded"`
}

// client data
type TomeDataClientAlias TomeData
type TomeDataClient struct {
	GuaranteedRarities      string        `json:"guaranteedRarities"`
	CardsRewarded           string        `json:"cardsRewarded"`

	*TomeDataClientAlias
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

	// server guarantees
	tome.GuaranteedRarities = util.StringToIntArray(client.GuaranteedRarities)

	// server rewards
	tome.CardsRewarded = util.StringToIntArray(client.CardsRewarded)

	return nil
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
		util.Must(err)

		// insert into table
		tomes[id] = &container.Tomes[i]
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