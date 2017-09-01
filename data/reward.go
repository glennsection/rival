package data

import (
	"encoding/json"
	"strconv"
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
	UseMultipliers 		bool 	 		`json:"useMultipliers,string"`	
	
	SpecificCards	 	[]DataId
	SpecificCounts 		[]int 			

	LegendaryCards 		int 			
	EpicCards 	 		int 			
	RareCards 	 		int 			
	RandomCards 		int			

	LegendaryChance 	float32 		
	EpicChance 			float32 		
	RareChance 			float32			

	LegendaryBounds 	[]int
	EpicBounds 			[]int
	RareBounds 			[]int
	CommonBounds 		[]int

	PremiumCurrency 	[]int 			
	StandardCurrency	[]int 			
}

type RewardDataClientAlias RewardData
type RewardDataClient struct {
	Type 				string 			`json:"rewardType"`	

	SpecificCards 		string 			`json:"specificCards"`
	SpecificCounts 		string 			`json:"specificCounts"`

	LegendaryCards 		string 			`json:"legendaryCards"`
	EpicCards 	 		string 			`json:"epicCards"`
	RareCards 	 		string 			`json:"rareCards"`
	RandomCards 		string 			`json:"randomCards"`

	LegendaryChance 	string 			`json:"legendaryChance"`
	EpicChance 			string 			`json:"epicChance"`
	RareChance 			string			`json:"rareChance"`

	LegendaryBounds 	string 			`json:"legendaryBounds"`
	EpicBounds 			string 			`json:"epicBounds"`
	RareBounds 			string 			`json:"rareBounds"`
	CommonBounds 		string 			`json:"commonBounds"`

	PremiumCurrency 	string 			`json:"premiumCurrency"`
	StandardCurrency 	string 			`json:"standardCurrency"`

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

	// specific rewards
	if client.SpecificCards != "" {
		specificCards := util.StringToStringArray(client.SpecificCards)

		reward.SpecificCards = make([]DataId, 0)
		for _, id := range specificCards {
			reward.SpecificCards = append(reward.SpecificCards, ToDataId(id))
		}
	}

	if client.SpecificCounts != "" {
		reward.SpecificCounts = util.StringToIntArray(client.SpecificCounts)

		if len(reward.SpecificCards) > len(reward.SpecificCounts) {
			for i := len(reward.SpecificCounts); i < len(reward.SpecificCards); i++ {
				reward.SpecificCounts = append(reward.SpecificCounts, 0)
			}
		}

		if len(reward.SpecificCards) < len(reward.SpecificCounts) {
			reward.SpecificCounts = reward.SpecificCounts[:len(reward.SpecificCards)]
		}
	}

	// card amounts  
	var i int64

	if client.LegendaryCards != "" {
		if i, err = strconv.ParseInt(client.LegendaryCards, 10, 64); err != nil { 
			panic(err) 
		} else { reward.LegendaryCards = int(i) }
	} 

	if client.EpicCards != "" {
		if i, err = strconv.ParseInt(client.EpicCards, 10, 64); err != nil { 
			panic(err) 
		} else { reward.EpicCards = int(i) }
	}

	if client.RareCards != "" {
		if i, err = strconv.ParseInt(client.RareCards, 10, 64); err != nil { 
			panic(err) 
		} else { reward.RareCards = int(i) }
	}

	if client.RandomCards != "" {
		if i, err = strconv.ParseInt(client.RandomCards, 10, 64); err != nil { 
			panic(err) 
		} else { reward.RandomCards = int(i) }
	}

	// bonus chances
	var f float64

	if client.LegendaryChance != "" {
		if f, err = strconv.ParseFloat(client.LegendaryChance, 64); err != nil { 
			panic(err) 
		} else { reward.LegendaryChance = float32(f) }
	}

	if client.EpicChance != "" {
		if f, err = strconv.ParseFloat(client.EpicChance, 64); err != nil { 
			panic(err) 
		} else { reward.EpicChance = float32(f) }
	}

	if client.RareChance != "" {
		if f, err = strconv.ParseFloat(client.RareChance, 64); err != nil { 
			panic(err) 
		} else { reward.RareChance = float32(f) }
	}

	// card bounds
	if reward.LegendaryCards > 0 || reward.RandomCards > 0 {
		if reward.LegendaryBounds = util.StringToIntArray(client.LegendaryBounds); len(reward.LegendaryBounds) != 2 {
			panic(errors.New(fmt.Sprintf("Improperly formatted bounds for Legendary cards in reward %s", reward.ID)))
		}
	} else { reward.LegendaryBounds = []int{0,0} }

	if reward.EpicCards > 0 || reward.RandomCards > 0 {
		if reward.EpicBounds = util.StringToIntArray(client.EpicBounds); len(reward.EpicBounds) != 2 {
			panic(errors.New(fmt.Sprintf("Improperly formatted bounds for Epic cards in reward %s", reward.ID)))
		}
	} else { reward.EpicBounds = []int{0,0} }

	if reward.RareCards > 0 || reward.RandomCards > 0 {
		if reward.RareBounds = util.StringToIntArray(client.RareBounds); len(reward.RareBounds) != 2 {
			panic(errors.New(fmt.Sprintf("Improperly formatted bounds for Rare cards in reward %s", reward.ID)))
		}
	} else { reward.RareBounds = []int{0,0} }

	if reward.RandomCards > 0 {
		if reward.CommonBounds = util.StringToIntArray(client.CommonBounds); len(reward.CommonBounds) != 2 {
			panic(errors.New(fmt.Sprintf("Improperly formatted bounds for Common cards in reward %s", reward.ID)))
		}
	} else { reward.CommonBounds = []int{0,0} }

	if client.PremiumCurrency != "" {
		if reward.PremiumCurrency = util.StringToIntArray(client.PremiumCurrency); len(reward.PremiumCurrency) != 2 {
			panic(errors.New(fmt.Sprintf("Improperly formatted bounds for Premium currency in reward %s", reward.ID)))
		}
	} else { reward.PremiumCurrency = []int{0,0} }

	if client.StandardCurrency != "" {
		if reward.StandardCurrency = util.StringToIntArray(client.StandardCurrency); len(reward.StandardCurrency) != 2 {
			panic(errors.New(fmt.Sprintf("Improperly formatted bounds for Standard currency in reward %s", reward.ID)))
		}
	} else { reward.StandardCurrency = []int{0,0} }

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

func (reward *RewardData)GetBoundsForPremiumCurrency() (int, int) { // helper for readability
	if len(reward.PremiumCurrency) != 2 { return 0, 0 }
	return reward.PremiumCurrency[0], reward.PremiumCurrency[1]
}

func (reward *RewardData)GetBoundsForStandardCurrency() (int, int) { // helper for readability
	if len(reward.StandardCurrency) != 2 { return 0, 0 }
	return reward.StandardCurrency[0], reward.StandardCurrency[1]
}

func (reward *RewardData)GetBoundsForRarity(rarity string) (int, int) { // helper for readability
	switch(rarity) {

	case "LEGENDARY":
		if len(reward.LegendaryBounds) != 2 { return 0,0 }
		return reward.LegendaryBounds[0], reward.LegendaryBounds[1]

	case "EPIC":
		if len(reward.EpicBounds) != 2 { return 0,0 }
		return reward.EpicBounds[0], reward.EpicBounds[1]

	case "RARE":
		if len(reward.RareBounds) != 2 { return 0,0 }
		return reward.RareBounds[0], reward.RareBounds[1]

	case "COMMON":
		if len(reward.CommonBounds) != 2 { return 0,0 }
		return reward.CommonBounds[0], reward.CommonBounds[1]
	}

	panic(errors.New("Invalid value passed as rarity"))
	return 0, 0
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