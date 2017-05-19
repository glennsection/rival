package controllers

import (
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/data"
)

func HandleCard() {
	HandleGameAPI("/card/upgrade", system.TokenAuthentication, UpgradeCard)
}

func UpgradeCard(context *util.Context) {
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
	player.XP += levelData.XP
	card.CardCount -= levelData.CardsNeeded
	card.Level += 1

	context.SetDirty([]int64{models.UpdateMask_XP, 
										 models.UpdateMask_Currency,
										 models.UpdateMask_Cards})
	context.Data = card

	player.Save(context.DB)
}