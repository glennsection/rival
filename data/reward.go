package data

import (
	"encoding/json"

	"bloodtales/util"
)

type RewardData struct {
	ID 					string 			`json:"id"`
	TomeIds 			[]DataId
	CardRarities		[]int 			
	CardAmounts			[]int
	MaxPremiumCurrency	int 			`json:"maxPremiumCurrency,string"`
	MinPremiumCurrency	int 			`json:"minPremiumCurrency,string"`
	MaxStandardCurrency	int 			`json:"maxStandardCurrency,string"`
	MinStandardCurrency	int 			`json:"minStandardCurrency,string"`
}

type RewardDataClientAlias RewardData
type RewardDataClient struct {
	Tomes 				string 			`json:"tomes"`		
	CardRarities 		string 			`json:"cardRarities"`
	CardAmounts 		string 			`json:"cardAmounts"`

	*RewardDataClientAlias
}

var rewards map[DataId]RewardData

type RewardDataParsed struct {
	Rewards []RewardData
} 

func (reward *RewardData)UnmarshalJSON(raw []byte) error { 
	// create client model
	client := &RewardDataClient {
		RewardDataClientAlias: (*RewardDataClientAlias)(reward),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw,client); err != nil {
		return err
	}

	// tomes
	reward.TomeIds = make([]DataId, 0)
	if client.Tomes != "" {
		tomes := util.StringToStringArray(client.Tomes)

		for _, tomeId := range tomes {
			reward.TomeIds = append(reward.TomeIds, ToDataId(tomeId))
		}
	} 

	// guaranteed card rarities
	if client.CardRarities != "" {
		reward.CardRarities = util.StringToIntArray(client.CardRarities)
	} else {
		reward.CardRarities = []int{0,0,0,0}
	}

	// card counts per rarity
	if client.CardAmounts != "" {
		reward.CardAmounts = util.StringToIntArray(client.CardAmounts)
	} else {
		reward.CardAmounts = []int{0,0,0,0}
	} 

	return nil
}

func LoadRewardData(raw []byte) {
	// parse and enter into system data
	container := &RewardDataParsed{}
	util.Must(json.Unmarshal(raw, container))

	rewards = map[DataId]RewardData {}
	for _,reward := range container.Rewards {
		id,err := mapDataName(reward.ID)
		util.Must(err)

		//insert into table
		rewards[id] = reward
	}

}

func GetRewardData(id DataId) *RewardData {
	reward := rewards[id]
	return &reward
}