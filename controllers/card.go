package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/data"
)

func HandleCard(application *system.Application) {
	application.HandleAPI("/card/upgrade", system.TokenAuthentication, UpgradeCard)
}

func UpgradeCard(context *system.Context) {
	id := context.Params.GetRequiredString("cardId")

	player := context.GetPlayer()
	cards := player.GetMapOfCardIndexes()
	index, valid := cards[data.ToDataId(id)]

	if !valid {
		context.Fail("Invalid ID")
		return
	}

	var card *models.Card = &player.Cards[index]
	levelData := data.GetCardProgressionData(card.GetData().Rarity, card.Level + 1)

	if player.StandardCurrency < levelData.Cost {
		context.Fail("Insufficient funds")
		return
	} else {
		if card.CardCount < levelData.CardsNeeded  {
			context.Fail("Requirements not met")
			return
		}
	}

	player.StandardCurrency -= levelData.Cost
	player.Xp += levelData.Xp
	card.CardCount -= levelData.CardsNeeded
	card.Level += 1

	player.Update(context.DB)
}