package models

import (
	"encoding/json"

	"bloodtales/util"
	"bloodtales/data"
)

//server model
type Reward struct {
	Data				*data.RewardData
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

	if reward.PremiumCurrency > 0 || reward.StandardCurrency > 0 {
		client["premiumCurrency"] = reward.PremiumCurrency
		client["standardCurrency"] = reward.StandardCurrency
	}

	if reward.OverflowCurrency > 0 {
		client["overflowAmounts"] = reward.OverflowAmounts
		client["overflowCurrency"] = reward.OverflowCurrency
	}

	return json.Marshal(client) 
}

func (player *Player)GetReward(rewardId data.DataId, league data.League, tier int) *Reward {
	rewardData := data.GetRewardData(rewardId)
	reward := CreateReward(rewardData, league, tier)
	player.checkForOverflow(reward)
	return reward
}

func (player *Player)GetRewards(rewardIds []data.DataId, league data.League, tier int) []*Reward {
	rewards := make([]*Reward, 0)

	for _, id := range rewardIds {
		rewards = append(rewards, player.GetReward(id, league, tier))
	}

	return rewards
}

func CreateCraftingReward(numCards int, rarity string) *Reward {
	reward := &Reward {
		Cards: make([]data.DataId, 0),
		NumRewarded: make([]int, 0),
	}

	possibleCards := data.GetCards(func(card *data.CardData) bool {
		return card.Rarity == rarity
	})

	for numCards > 0 {
		// select a random character card
		index := util.RandomIntn(len(possibleCards))
		card := possibleCards[index]

		// add the character card to the reward
		reward.Cards = append(reward.Cards, card)
		reward.NumRewarded = append(reward.NumRewarded, 1)

		numCards--
	}

	return reward
}

func CreateReward(rewardData *data.RewardData, league data.League, tier int) *Reward {
	reward := &Reward{
		Data: rewardData,
		ItemID: rewardData.ItemID,
		Type: rewardData.Type,
	}

	volumeMultiplier := 1.0
	standardCurrencyMultiplier := 1.0
	premiumCurrencyMultiplier := 1.0

	if(rewardData.UseMultipliers) {
		volumeMultiplier = data.GetLeagueData(league).TomeVolumeMultiplier
		standardCurrencyMultiplier = data.GetLeagueData(league).StandardCurrencyMultiplier
		premiumCurrencyMultiplier = data.GetLeagueData(league).PremiumCurrencyMultiplier
	}
	
	reward.getCurrencyRewards(rewardData, standardCurrencyMultiplier, premiumCurrencyMultiplier)
	reward.getCardRewards(rewardData, tier, volumeMultiplier)

	return reward
}

func (reward *Reward)getCurrencyRewards(rewardData *data.RewardData, standardMultiplier float64, premiumMultiplier float64) {

	minPremiumCurrency, maxPremiumCurrency := rewardData.GetBoundsForPremiumCurrency()
	minStandardCurrency, maxStandardCurrency := rewardData.GetBoundsForStandardCurrency()

	minStandardCurrency = int(float64(minStandardCurrency) * standardMultiplier)
	maxStandardCurrency = int(float64(maxStandardCurrency) * standardMultiplier)
	minPremiumCurrency = int(float64(minPremiumCurrency) * premiumMultiplier)
	maxPremiumCurrency = int(float64(maxPremiumCurrency) * premiumMultiplier)

	if maxPremiumCurrency == minPremiumCurrency {
		reward.PremiumCurrency = maxPremiumCurrency
	} else {
		reward.PremiumCurrency = minPremiumCurrency + util.RandomIntn(maxPremiumCurrency - minPremiumCurrency + 1)
	}
	
	if maxStandardCurrency == minStandardCurrency {
		reward.StandardCurrency = maxStandardCurrency
	} else {
		reward.StandardCurrency = minStandardCurrency + util.RandomIntn(maxStandardCurrency - minStandardCurrency + 1)
	}
}

func (reward *Reward)getCardRewards(rewardData *data.RewardData, tier int, volumeMultiplier float64) {
	reward.Cards = make([]data.DataId, 0)
	reward.NumRewarded = make([]int, 0)

	// first assign cards for the guaranteed rarities
	reward.getCardsForRarity(rewardData, "LEGENDARY", int(float64(rewardData.LegendaryCards) * volumeMultiplier), tier, volumeMultiplier)
	reward.getCardsForRarity(rewardData, "EPIC", int(float64(rewardData.EpicCards) * volumeMultiplier), tier, volumeMultiplier)
	reward.getCardsForRarity(rewardData, "RARE", int(float64(rewardData.RareCards) * volumeMultiplier), tier, volumeMultiplier)

	// next roll for a rare or better card
	remainingCards := int(float64(rewardData.RandomCards) * volumeMultiplier)
	reward.rollForCard(rewardData, &remainingCards, tier, volumeMultiplier)

	//finally, fill out the remaining cards
	reward.getCardsForRarity(rewardData, "COMMON", remainingCards, tier, volumeMultiplier)
}

func (reward *Reward)getCardsForRarity(rewardData *data.RewardData, rarity string, numCards int, tier int, volumeMultiplier float64) {
	if numCards == 0 {
		return
	}

	lowerBound, upperBound := rewardData.GetBoundsForRarity(rarity)
	lowerBound = int(float64(lowerBound) * volumeMultiplier)
	upperBound = int(float64(upperBound) * volumeMultiplier)

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

func (reward *Reward)rollForCard(rewardData *data.RewardData, remainingCards *int, tier int, volumeMultiplier float64) {
	roll := float32(util.RandomIntn(10000)) / 100
	rarity := ""

	var possibleCards []data.DataId

	getCardsFunc := func(card *data.CardData) bool {
		id := data.ToDataId(card.Name)
		for _, cardId := range reward.Cards {
			if id == cardId { // ensure we don't already have this card
				return false
			}
		}

		return card.Rarity == rarity && card.Tier <= tier
	}

	//first determine if we rolled successfully
	if roll <= rewardData.LegendaryChance {
		rarity = "LEGENDARY"
		possibleCards = data.GetCards(getCardsFunc)
	}
	if len(possibleCards) == 0 && roll <= (rewardData.EpicChance + rewardData.LegendaryChance) {
		rarity = "EPIC"
		possibleCards = data.GetCards(getCardsFunc)
	}
	if len(possibleCards) == 0 && roll <= (rewardData.RareChance + rewardData.EpicChance + rewardData.LegendaryChance) {
		rarity = "RARE"
		possibleCards = data.GetCards(getCardsFunc)
	}	
	if len(possibleCards) == 0 {
		return
	}

	lowerBound, upperBound := rewardData.GetBoundsForRarity(rarity)
	lowerBound = int(float64(lowerBound) * volumeMultiplier)
	upperBound = int(float64(upperBound) * volumeMultiplier)

	if upperBound > *remainingCards {
			upperBound = *remainingCards
	}

	*remainingCards -= reward.getCard(&possibleCards, lowerBound, upperBound)
}

func (reward *Reward)getCard(possibleCards *[]data.DataId, lowerBound int, upperBound int) (cardsRewarded int) {
	// select a random character card
	index := util.RandomIntn(len(*possibleCards))
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
		cardsRewarded = util.RandomRange(lowerBound, upperBound)
	}

	// add the character card to the reward
	reward.Cards = append(reward.Cards, card)
	reward.NumRewarded = append(reward.NumRewarded, cardsRewarded)

	return
}

func (player *Player)checkForOverflow(reward *Reward) {
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
		if card := player.GetCard(id); card != nil && (card.CardCount + reward.NumRewarded[index]) >= maxCards {
			return card.CardCount + reward.NumRewarded[index] - maxCards
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