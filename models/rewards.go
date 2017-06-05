package models

import (
	"encoding/json"

	"bloodtales/data"
)

//server model
type Reward struct {
	Tomes 				[]data.DataId
	Cards 				[]data.DataId
	NumRewarded			[]int 			
	PremiumCurrency 	int 			
	StandardCurrency 	int 			
}

//client model
type RewardClient struct {
	Tomes 				[]data.TomeData	`json:tomes,omitempty`
	Cards 				[]data.CardData	`json:cards,omitempty`
	NumRewarded			[]int 			`json:numRewarded,omitempty` 			
	PremiumCurrency 	int 			`json:PremiumCurrency,omitempty`		
	StandardCurrency 	int 			`json:StandardCurrency,omitempty`
}

// custom marshalling
func (reward *Reward) MarshalJSON() ([]byte, error) {
	client := map[string]interface{}{}

	// tomes
	if len(reward.Tomes) > 0 {
		tomes := make([]data.TomeData, len(reward.Tomes))

		for i, id := range reward.Tomes {
			tomes[i] = *(data.GetTome(id))
		}

		client["tomes"] = tomes	
	}

	// cards
	if len(reward.Cards) > 0 {
		cards := make([]data.CardData, len(reward.Cards))

		for i, id := range reward.Cards {
			cards[i] = *(data.GetCard(id))
		}

		client["cards"] = cards
		client["numRewarded"] = reward.NumRewarded
	}

	if reward.PremiumCurrency > 0 {
		client["premiumCurrency"] = reward.PremiumCurrency
		client["standardCurrency"] = reward.StandardCurrency
	}

	return json.Marshal(client)
}