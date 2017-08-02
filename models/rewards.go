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
	ItemID 				string
	Type 				data.RewardType
	Cards 				[]data.DataId
	NumRewarded			[]int 			
	PremiumCurrency 	int 			
	StandardCurrency 	int 
	OverflowAmounts 	[]int
	OverflowCurrency 	int		
}

// custom marshalling
func (reward *Reward) MarshalJSON() ([]byte, error) {
	client := map[string]interface{}{}

	client["itemId"] = reward.ItemID

	var err error
	err = nil

	if client["rewardType"], err = data.RewardTypeToString(reward.Type); err != nil {
		panic(err)
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

	if reward.OverflowCurrency > 0 {
		client["overflowAmounts"] = reward.OverflowAmounts
		client["overflowCurrency"] = reward.OverflowCurrency
	}

	return json.Marshal(client) 
}

func (player *Player) GetReward(rewardId data.DataId) *Reward {
	rewardData := data.GetRewardData(rewardId)
	return player.CreateReward(rewardData, false)
}

func (player *Player) GetRewards(rewardIds []data.DataId) []*Reward {
	rewards := make([]*Reward, 0)

	for _, id := range rewardIds {
		rewards = append(rewards, player.GetReward(id))
	}

	return rewards
}

func (player *Player) CreateCraftingReward(numCards int, rarity string) *Reward {
	var rarities []int
	var numRewarded []int

	switch(rarity) {
	case "COMMON":
		rarities = []int{numCards,0,0,0}
		numRewarded = []int{numCards,0,0,0}

	case "RARE":
		rarities = []int{0,numCards,0,0}
		numRewarded = []int{0,numCards,0,0}

	case "EPIC":
		rarities = []int{0,0,numCards,0}
		numRewarded = []int{0,0,numCards,0}

	case "LEGENDARY":
		rarities = []int{0,0,0,numCards}
		numRewarded = []int{0,0,0,numCards}
	}

	rewardData := &data.RewardData {
		Type: data.RewardType_Card,
		CardRarities: rarities,
		CardAmounts: numRewarded,
	}

	return player.CreateReward(rewardData, true)
}

func (player *Player)CreateReward(rewardData *data.RewardData, allowDuplicates bool) *Reward {
	reward := &Reward{
		ItemID: rewardData.ItemID,
		Type: rewardData.Type,
	}
	
	reward.getCurrencyRewards(rewardData)
	reward.getCardRewards(rewardData, player.GetLevel(), allowDuplicates)
	reward.getOverflowAmounts(player)

	return reward
}

func (reward *Reward)getCurrencyRewards(rewardData *data.RewardData) {
	rand.Seed(time.Now().UTC().UnixNano())

	if rewardData.MaxPremiumCurrency == rewardData.MinPremiumCurrency {
		reward.PremiumCurrency = rewardData.MaxPremiumCurrency
	} else {
		reward.PremiumCurrency = rewardData.MinPremiumCurrency + rand.Intn(rewardData.MaxPremiumCurrency - rewardData.MinPremiumCurrency + 1)
	}
	
	if rewardData.MaxStandardCurrency == rewardData.MinStandardCurrency {
		reward.StandardCurrency = rewardData.MaxStandardCurrency
	} else {
		reward.StandardCurrency = rewardData.MinStandardCurrency + rand.Intn(rewardData.MaxStandardCurrency - rewardData.MinStandardCurrency + 1)
	}
}

func (reward *Reward)getCardRewards(rewardData *data.RewardData, tier int, allowDuplicates bool) {
	rarities := []string{"COMMON","RARE","EPIC","LEGENDARY"}
	reward.Cards = make([]data.DataId, 0)
	reward.NumRewarded = make([]int, 0)

	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < len(rewardData.CardRarities); i++ {

		totalCards := rewardData.CardAmounts[i]
		cardInstances := rewardData.CardRarities[i]

		if totalCards == 0 || cardInstances == 0 {
			continue;
		}


		getCards := func(card *data.CardData) bool {
			return card.Rarity == rarities[i] && card.Tier <= tier
		}

		cardSlice := data.GetCards(getCards)

		cardCountFloor := totalCards / (cardInstances * 2)
		if cardCountFloor == 0 {
			cardCountFloor = 1
		}

		for j := 0; j < cardInstances; j++ {
			if len(cardSlice) == 0 {
				break
			}

			index := rand.Intn(len(cardSlice))

			card := cardSlice[index]

			if !allowDuplicates {
				if index != (len(cardSlice) - 1) {
					cardSlice[index] = cardSlice[len(cardSlice) - 1]
				} 
				cardSlice = cardSlice[:len(cardSlice) - 1]
			}

			var cardsRewarded int
			if j == cardInstances - 1 {
				cardsRewarded = totalCards
			} else {
				randMax := totalCards - (cardCountFloor * (cardInstances - j))

				if randMax > 0 {
					cardsRewarded = rand.Intn(randMax) + cardCountFloor
				} else {
					cardsRewarded = cardCountFloor
				}

				totalCards -= cardsRewarded
			}

			reward.Cards = append(reward.Cards, card)
			reward.NumRewarded = append(reward.NumRewarded, cardsRewarded)
		}
	}
}

func (reward *Reward)getOverflowAmounts(player *Player) {
	reward.OverflowAmounts = make([]int, len(reward.Cards))

	for i,_ := range reward.OverflowAmounts {
		overflow := reward.getOverflowForIndex(player, i)
		reward.OverflowAmounts[i] = overflow
		reward.OverflowCurrency += overflow * data.Config().LegendaryCardCurrencyValue
	}
}

//since legendary cards can't be used for crafting, any cards over their max obtainable amount should be converted
//into standard currency. this function determines the amount of cards that overflow by index.
func (reward *Reward)getOverflowForIndex(player *Player, index int) int {
	id := reward.Cards[index]
	rarity := data.GetCard(id).Rarity

	if rarity == "LEGENDARY" {
		maxCards := data.GetMaxCardCount(rarity)
		if cardRef,hasCard := player.HasCard(id); hasCard && (cardRef.CardCount + reward.NumRewarded[index]) >= maxCards {
			return cardRef.CardCount + reward.NumRewarded[index] - maxCards
		} 
	}

	return 0
}

// player functions below

func (player *Player) AddRewards(reward *Reward, context *util.Context) (err error) {
	player.PremiumCurrency += reward.PremiumCurrency
	player.StandardCurrency += reward.StandardCurrency + reward.OverflowCurrency

	for i, id := range reward.Cards {
		overflow := reward.getOverflowForIndex(player, i) 
		player.AddCards(id, (reward.NumRewarded[i] - overflow))
	}

	if context != nil {
		err = player.Save(context)
	}
	return
}