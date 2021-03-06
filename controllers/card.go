package controllers

import (
	"fmt"
	"time"
	"encoding/json"
	
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/data"
)

func handleCard() {
	handleGameAPI("/card/upgrade", system.TokenAuthentication, UpgradeCard)
	handleGameAPI("/card/craft", system.TokenAuthentication, CraftCard)
	handleGameAPI("/card/viewed", system.TokenAuthentication, OnCardViewed)
}

func UpgradeCard(context *util.Context) {
	// parse parameters
	id := context.Params.GetRequiredString("cardId")

	player := GetPlayer(context)
	cardIndexes := player.GetMapOfCardIndexes()
	index, valid := cardIndexes[data.ToDataId(id)]
	if !valid {
		context.Fail("Invalid ID")
		return
	}

	card := &player.Cards[index]
	cardData := card.GetData()
	maxLevel := data.GetMaxLevel(cardData.Rarity)

	if card.Level == maxLevel {
		context.Fail("Cannot upgrade any further")
		return
	}

	levelData := data.GetCardProgressionData(card.GetData().Rarity, card.Level)

	if player.StandardCurrency < levelData.Cost {
		context.Fail("Insufficient funds")
		return
	} 

	if (card.CardCount < levelData.CardsNeeded) || 
	(card.Level == (maxLevel - 1) && (card.WinCount < cardData.AwakenGamesNeeded || card.LeaderWinCount < cardData.AwakenLeaderGamesNeeded)) {
		context.Fail("Requirements not met")
		return
	}

	player.StandardCurrency -= levelData.Cost

	// get previous player level
	previousLevel := player.GetLevel()

	// add XP
	player.XP += levelData.XP

	// check level-up
	currentLevel := player.GetLevel()
	if previousLevel != currentLevel {
	// analytics
		if util.HasSQLDatabase() {
			InsertTrackingSQL(context, "levelUp", 0, "", "", currentLevel, 0, nil)
		} else {
			InsertTracking(context, "levelUp", bson.M { "level": currentLevel }, 0)
		}
	}

	card.CardCount -= levelData.CardsNeeded
	card.Level += 1

	// analytics
	if util.HasSQLDatabase() {
		InsertTrackingSQL(context, "cardLevelUp", 0, data.ToDataName(card.DataID), "", card.Level, float64(levelData.Cost), nil)
	} else {
		InsertTracking(context, "cardLevelUp", bson.M { "cardId": data.ToDataName(card.DataID), 
													"level": card.Level,
													"price": levelData.Cost }, 0)
	}

	player.SetDirty(models.PlayerDataMask_Cards, models.PlayerDataMask_Currency, models.PlayerDataMask_XP)
	player.Save(context)

	context.SetData("card", card)
}

func CraftCard(context *util.Context) {
	// parse parameters
	rarity := context.Params.GetRequiredString("rarity")
	cardsJs := context.Params.GetRequiredString("cards")

	// validate the query
	var consumableCards map[string]int
	json.Unmarshal([]byte(cardsJs), &consumableCards)
	if len(consumableCards) == 0 {
		context.Fail("Malformed Request")
		return
	}

	player := GetPlayer(context)
	baseCost := data.GetCraftingCost(rarity)
	cardsNeeded := data.GetCraftingXpNeeded(rarity)

	cardsSupplied := 0
	for cardId, num := range consumableCards {
		dataId := data.ToDataId(cardId)
		if card := player.GetCard(dataId); card != nil && card.CardCount >= num {
			cardsSupplied += num
			// deduct the cards supplied - if we fail later, we won't update the db and the change wont stick
			card.CardCount -= num
		} else {
				context.Fail(fmt.Sprintf("Insufficient cards of type %s", cardId))
				return
		}
	}

	fmt.Println(fmt.Sprintf("Cards Needed: %d", cardsNeeded))

	// final validation step: make sure the user supplied the correct amount of cards and can afford the exchange
	if cardsSupplied == 0 || cardsSupplied % cardsNeeded != 0 {
		context.Fail("Insufficient Cards")
		return
	} else {
		if player.StandardCurrency < baseCost {
			context.Fail("Insufficient Funds")
			return
		}
	}

	// subtract the cost of the transaction and add (cardsSupplied/cost) random cards
	numCards := cardsSupplied / cardsNeeded
	player.StandardCurrency -= baseCost * numCards

	reward := models.CreateCraftingReward(numCards, rarity)
	player.AddRewards(reward, nil)

	cardsGained := map[string]int{} // used for analytics
	//range over the cards returned in the reward object so we can group together duplicates
	for _,id := range reward.Cards {
		name := data.ToDataName(id)
		if _,ok := cardsGained[name]; ok {
			cardsGained[name]++
			continue
		}
		cardsGained[name] = 1
	}

	// analytics
	currentTime := util.TimeToTicks(time.Now().UTC())
	
	for cardId, num := range consumableCards {
		if util.HasSQLDatabase() {
			InsertTrackingSQL(context, "cardConsumed", currentTime, cardId, "", num, 0, nil)
		} else {
			InsertTracking(context, "cardConsumed", bson.M { "time":currentTime,
														 "cardId":cardId,
														 "count":num }, 0)
		}
	}

	for cardId, num := range cardsGained {
		if util.HasSQLDatabase() {
			InsertTrackingSQL(context, "cardCrafted", currentTime, cardId, "", num, float64(baseCost * num), nil)
		} else {
			InsertTracking(context, "cardCrafted", bson.M { "time":currentTime,
														"cardId":cardId,
														"count":num,
														"goldSpent":(baseCost * num) }, 0)
		}
	} 

	if util.HasSQLDatabase() {
		TrackRewardsSQL(context, reward, currentTime)
	}else{
		TrackRewards(context, reward)
	}

	player.Save(context)
	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
	context.SetData("reward", reward)
}

func OnCardViewed(context *util.Context) {
	// parse parameters
	id := context.Params.GetRequiredString("cardId")
	
	player := GetPlayer(context)
	card := player.GetCard(data.ToDataId(id))

	card.HasInteractedWith = true;

	player.Save(context)
	//player.SetDirty(models.PlayerDataMask_Cards) //Don't do for now
}