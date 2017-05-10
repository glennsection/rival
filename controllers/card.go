package controllers

import (
	"bloodtales/system"
	"bloodtales/data"
)

func HandleCard(application *system.Application) {
	application.HandleAPI("/card/upgrade", system.TokenAuthentication, UpgradeCard)
}

func UpgradeCard(context *system.Context) {
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

	player.StandardCurrency -= levelData.Cost
	player.Xp += levelData.Xp
	player.Level = data.GetAccountLevel(player.Xp)
	card.CardCount -= levelData.CardsNeeded
	card.Level += 1

	context.Data = card
	player.Update(context.DB)
}