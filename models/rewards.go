package models

import (
	"encoding/json"
	"math/rand"
	"time"

	"bloodtales/util"
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
	Tomes 				[]string		`json:tomes,omitempty`
	Cards 				[]string		`json:cards,omitempty`
	NumRewarded			[]int 			`json:numRewarded,omitempty` 			
	PremiumCurrency 	int 			`json:PremiumCurrency,omitempty`		
	StandardCurrency 	int 			`json:StandardCurrency,omitempty` 
}

// custom marshalling
func (reward *Reward) MarshalJSON() ([]byte, error) {
	client := map[string]interface{}{}

	// tomes
	if len(reward.Tomes) > 0 {
		tomes := make([]string, len(reward.Tomes))

		for i, id := range reward.Tomes {
			tomes[i] = data.ToDataName(id)
		}

		client["tomes"] = tomes	
	}

	// cards
	if len(reward.Cards) > 0 {
		cards := make([]string, len(reward.Cards))

		for i, id := range reward.Cards {
			cards[i] = data.ToDataName(id)
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

func GetReward(rewardId data.DataId, tier int) *Reward {
	rewardData := data.GetRewardData(rewardId)
	reward := &Reward{}
	
	reward.getCurrencyRewards(rewardData)
	reward.getCardRewards(rewardData, tier)
	reward.getTomeRewards(rewardData)

	return reward
}

func (reward *Reward)getCurrencyRewards(rewardData *data.RewardData) {
	rand.Seed(time.Now().UTC().UnixNano())

	if rewardData.MaxPremiumCurrency == rewardData.MinPremiumCurrency {
		reward.PremiumCurrency = rewardData.MaxPremiumCurrency
	} else {
		reward.PremiumCurrency = rewardData.MinPremiumCurrency + rand.Intn(rewardData.MaxPremiumCurrency - rewardData.MinPremiumCurrency)
	}
	
	if rewardData.MaxStandardCurrency == rewardData.MinStandardCurrency {
		reward.StandardCurrency = rewardData.MaxStandardCurrency
	} else {
		reward.StandardCurrency = rewardData.MinStandardCurrency + rand.Intn(rewardData.MaxStandardCurrency - rewardData.MinStandardCurrency)
	}
}

func (reward *Reward)getCardRewards(rewardData *data.RewardData, tier int) {
	rarities := []string{"COMMON","RARE","EPIC","LEGENDARY"}
	reward.Cards = make([]data.DataId, 0)
	reward.NumRewarded = make([]int, 0)

	for i := 0; i < len(rewardData.CardRarities); i++ {
		getCards := func(card *data.CardData) bool {
			return card.Rarity == rarities[i] && card.Tier <= tier
		}

		cardSlice := data.GetCards(getCards)

		for j := 0; j < rewardData.CardRarities[i]; j++ {
			if len(cardSlice) == 0 {
				break
			}

			rand.Seed(time.Now().UTC().UnixNano())
			index := rand.Intn(len(cardSlice))

			card := cardSlice[index]

			if index != (len(cardSlice) - 1) {
				cardSlice[index] = cardSlice[len(cardSlice) - 1]
			} 
			cardSlice = cardSlice[:len(cardSlice) - 1]

			reward.Cards = append(reward.Cards, card)
			reward.NumRewarded = append(reward.NumRewarded, rewardData.CardAmounts[i])
		}
	}
}

func (reward *Reward)getTomeRewards(rewardData *data.RewardData) {
	reward.Tomes = rewardData.TomeIds
}

// player functions below

func (player *Player) AddRewards(reward *Reward, context *util.Context) (err error) {
	player.PremiumCurrency += reward.PremiumCurrency
	player.StandardCurrency += reward.StandardCurrency

	for i, id := range reward.Cards {
		player.AddCards(id, reward.NumRewarded[i])
	}

	var tome *Tome
	for _, id := range reward.Tomes {
		//if the player has an open tome slot, add this tome
		tome = nil
		var index int
		for i, tomeSlot := range player.Tomes {
			if tomeSlot.State == TomeEmpty {
				index = i
				tome = &player.Tomes[i]
				break
			}
		}
		if tome != nil {
			player.Tomes[index].DataID = id
			player.Tomes[index].State = TomeLocked
			player.Tomes[index].UnlockTime = util.TicksToTime(0)
		}
	}

	if context != nil {
		err = player.Save(context)
	}
	return
}