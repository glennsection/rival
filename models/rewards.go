package models

import (
	"fmt"
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
	fmt.Println(fmt.Sprintf("Cards in reward: %d", len(reward.Cards)))
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
	return player.CreateReward(rewardData)
}

func (player *Player) GetRewards(rewardIds []data.DataId) []*Reward {
	rewards := make([]*Reward, 0)

	for _, id := range rewardIds {
		rewards = append(rewards, player.GetReward(id))
	}

	return rewards
}

func (player *Player) CreateCraftingReward(numCards int, rarity string) *Reward {
	reward := &Reward {
		Cards: make([]data.DataId, 0),
		NumRewarded: make([]int, 0),
	}

	rand.Seed(time.Now().UTC().UnixNano())

	tier := player.GetLevel()

	possibleCards := data.GetCards(func(card *data.CardData) bool {
		return card.Rarity == rarity && card.Tier <= tier
	})

	for numCards > 0 {
		// select a random character card
		index := rand.Intn(len(possibleCards))
		card := possibleCards[index]

		// add the character card to the reward
		reward.Cards = append(reward.Cards, card)
		reward.NumRewarded = append(reward.NumRewarded, 1)

		numCards--
	}

	return reward
}

func (player *Player)CreateReward(rewardData *data.RewardData) *Reward {
	reward := &Reward{
		ItemID: rewardData.ItemID,
		Type: rewardData.Type,
	}
	
	reward.getCurrencyRewards(rewardData)
	reward.getCardRewards(rewardData, player.GetLevel())
	reward.getOverflowAmounts(player)

	return reward
}

func (reward *Reward)getCurrencyRewards(rewardData *data.RewardData) {

	minPremiumCurrency, maxPremiumCurrency := rewardData.GetBoundsForPremiumCurrency()
	minStandardCurrency, maxStandardCurrency := rewardData.GetBoundsForStandardCurrency()

	rand.Seed(time.Now().UTC().UnixNano())



	if maxPremiumCurrency == minPremiumCurrency {
		reward.PremiumCurrency = maxPremiumCurrency
	} else {
		reward.PremiumCurrency = minPremiumCurrency + rand.Intn(maxPremiumCurrency - minPremiumCurrency + 1)
	}
	
	if maxStandardCurrency == minStandardCurrency {
		reward.StandardCurrency = maxStandardCurrency
	} else {
		reward.StandardCurrency = minStandardCurrency + rand.Intn(maxStandardCurrency - minStandardCurrency + 1)
	}
}

func (reward *Reward)getCardRewards(rewardData *data.RewardData, tier int) {
	reward.Cards = make([]data.DataId, 0)
	reward.NumRewarded = make([]int, 0)

	rand.Seed(time.Now().UTC().UnixNano())

	// first assign cards for the guaranteed rarities
	reward.getCardsForRarity(rewardData, "LEGENDARY", rewardData.LegendaryCards, tier)
	reward.getCardsForRarity(rewardData, "EPIC", rewardData.EpicCards, tier)
	reward.getCardsForRarity(rewardData, "RARE", rewardData.RareCards, tier)

	// next roll for a rare or better card
	remainingCards := rewardData.RandomCards
	reward.rollForCard(rewardData, &remainingCards, tier)

	//finally, fill out the remaining cards
	reward.getCardsForRarity(rewardData, "COMMON", remainingCards, tier)
}

func (reward *Reward)getCardsForRarity(rewardData *data.RewardData, rarity string, numCards int, tier int) {
	if numCards == 0 {
		return
	}

	lowerBound, upperBound := rewardData.GetBoundsForRarity(rarity)

	possibleCards := data.GetCards(func(card *data.CardData) bool {
		return card.Rarity == rarity && card.Tier <= tier
	})

	startingIndex := len(reward.Cards)

	for numCards > 0 {
		if len(possibleCards) == 0 {
			return
		}

		if upperBound >= numCards {
			lowerBound = numCards
			upperBound = numCards
		}

		cardsRewarded := reward.getCard(&possibleCards, lowerBound, upperBound)
		numCards -= cardsRewarded

		if numCards != 0 && numCards < lowerBound {
			for i := len(reward.NumRewarded) - 1; i >= startingIndex && numCards > 0; i-- {
				diff := upperBound - reward.NumRewarded[i]

				if diff > numCards {
					reward.NumRewarded[i] += numCards
					numCards = 0
				} else {
					reward.NumRewarded[i] += diff
					numCards -= diff
				}
			}
		}
	}
}


func (reward *Reward)rollForCard(rewardData *data.RewardData, remainingCards *int, tier int) {
	roll := float32(rand.Intn(100)) + rand.Float32()
	rarity := ""

	//first determine if we rolled successfully
	if roll <= (rewardData.RareChance + rewardData.EpicChance + rewardData.LegendaryChance) {
		rarity = "RARE"
	}	
	if roll <= (rewardData.EpicChance + rewardData.LegendaryChance) {
		rarity = "EPIC"
	}
	if roll <= rewardData.LegendaryChance {
		rarity = "LEGENDARY"
	}
	if rarity == "" {
		return
	}

	possibleCards := data.GetCards(func(card *data.CardData) bool {
		id := data.ToDataId(card.Name)
		for _, cardId := range reward.Cards {
			if id == cardId { // ensure we don't already have this card
				return false
			}
		}

		return card.Rarity == rarity && card.Tier <= tier
	})

	if len(possibleCards) == 0 {
		return
	}

	lowerBound, upperBound := rewardData.GetBoundsForRarity(rarity)
	if upperBound > *remainingCards {
			upperBound = *remainingCards
	}

	*remainingCards -= reward.getCard(&possibleCards, lowerBound, upperBound)
}

func (reward *Reward)getCard(possibleCards *[]data.DataId, lowerBound int, upperBound int) (cardsRewarded int) {
	// select a random character card
	index := rand.Intn(len(*possibleCards))
	card := (*possibleCards)[index]

	// remove that card from the slice of possible character cards
	if index != (len(*possibleCards) - 1) {
		(*possibleCards)[index] = (*possibleCards)[len(*possibleCards) - 1]
	} 
	*possibleCards = (*possibleCards)[:len(*possibleCards) - 1]

	// determine the card count for selected character card
	if upperBound <= lowerBound {
		cardsRewarded = upperBound
	} else {
		cardsRewarded = rand.Intn(upperBound - lowerBound + 1) + lowerBound
	}

	// add the character card to the reward
	reward.Cards = append(reward.Cards, card)
	reward.NumRewarded = append(reward.NumRewarded, cardsRewarded)

	return
}

func (reward *Reward)getOverflowAmounts(player *Player) {
	reward.OverflowAmounts = make([]int, len(reward.Cards))

	for i,_ := range reward.OverflowAmounts {
		overflow := reward.getOverflowForIndex(player, i)
		reward.OverflowAmounts[i] = overflow
		reward.OverflowCurrency += overflow * data.GameplayConfig.LegendaryCardCurrencyValue
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