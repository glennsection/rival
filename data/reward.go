package data

import (
	"encoding/json"
	"errors"
	"fmt"

	"bloodtales/util"
)

type RewardType int
const (
	RewardType_StandardCurrency RewardType = iota
	RewardType_PremiumCurrency
	RewardType_Card
	RewardType_Tome
)


type RewardData struct {
	ID 					string 			`json:"id"`
	ItemID 				string 			`json:"itemId"`
	Type 				RewardType 		
	CardRarities		[]int 			
	CardAmounts			[]int
	MaxPremiumCurrency	int 			`json:"maxPremiumCurrency,string"`
	MinPremiumCurrency	int 			`json:"minPremiumCurrency,string"`
	MaxStandardCurrency	int 			`json:"maxStandardCurrency,string"`
	MinStandardCurrency	int 			`json:"minStandardCurrency,string"`
}

type RewardDataClientAlias RewardData
type RewardDataClient struct {
	Type 				string 			`json:"rewardType"`	
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

	var err error
	err = nil

	// unmarshal to client model
	if err = json.Unmarshal(raw,client); err != nil {
		return err
	}

	// type
	if reward.Type, err = StringToRewardType(client.Type); err != nil {
		return err
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

func RewardTypeToString(val RewardType) (string, error) {
	switch val {
	case RewardType_StandardCurrency:
		return "StandardCurrency", nil
	case RewardType_PremiumCurrency:
		return "PremiumCurrency", nil
	case RewardType_Tome:
		return "Tome", nil
	case RewardType_Card:
		return "Card", nil
	}
	
	return "", errors.New("Invalid value passed as RewardType")
}

func StringToRewardType(val string) (RewardType, error) {
	switch val {
	case "StandardCurrency":
		return RewardType_StandardCurrency, nil
	case "PremiumCurrency":
		return RewardType_PremiumCurrency, nil
	case "Tome":
		return RewardType_Tome, nil
	case "Card":
		return RewardType_Card, nil
	}

	return RewardType_StandardCurrency, errors.New(fmt.Sprintf("Cannot convert %s to RewardType", val))
}