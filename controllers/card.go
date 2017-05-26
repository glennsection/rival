package controllers

import (
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/data"
)

func handleCard() {
	handleGameAPI("/card/upgrade", system.TokenAuthentication, UpgradeCard)
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
	levelData := data.GetCardProgressionData(card.GetData().Rarity, card.Level)

	if player.StandardCurrency < levelData.Cost {
		context.Fail("Insufficient funds")
		return
	} else {
		if card.CardCount < levelData.CardsNeeded  {
			context.Fail("Requirements not met")
			return
		}
	}

	// get previous player level
	previousLevel := data.GetAccountLevel(player.XP)

	player.StandardCurrency -= levelData.Cost
	player.XP += levelData.XP

	// check level-up
	currentLevel := data.GetAccountLevel(player.XP)
	if previousLevel != currentLevel {
		InsertTracking(context, "levelUp", bson.M { "level": currentLevel }, 0)
	}

	card.CardCount -= levelData.CardsNeeded
	card.Level += 1

	player.SetDirty(models.PlayerDataMask_All)
	player.Save(context.DB)

	context.SetData("card", card)
}