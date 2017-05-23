package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/data"
	"bloodtales/util"
)

func HandlePurchase() {
	HandleGameAPI("/purchase", system.TokenAuthentication, Purchase)
}

func Purchase(context *util.Context) {
	// parse parameters
	itemId := context.Params.GetRequiredString("itemId")

	// get store item
	storeItem := data.GetStoreItem(data.ToDataId(itemId))
	if storeItem == nil {
		context.Fail("Invalid store purchase")
		return
	}

	// get player
	player := GetPlayer(context)

	// check store item currency cost
	switch storeItem.Currency {

	case data.CurrencyReal:
		// TODO - verify mobile store receipt

	case data.CurrencyPremium:
		cost := int(storeItem.Cost)
		if player.PremiumCurrency < cost {
			context.Fail("Insufficient funds")
			return
		}
		player.PremiumCurrency -= cost

	}

	// handle store item category
	switch storeItem.Category {

	case data.StoreCategoryPremiumCurrency:
		player.PremiumCurrency += storeItem.Quantity
		player.SetDirty(models.PlayerDataMask_Currency)

	case data.StoreCategoryStandardCurrency:
		player.StandardCurrency += storeItem.Quantity
		player.SetDirty(models.PlayerDataMask_Currency)

	case data.StoreCategoryTomes:
		// claim tome
		tomeId := storeItem.ItemID
		reward, err := player.ClaimTome(context.DB, tomeId)
		util.Must(err)
		
		if reward == nil {
			context.Fail("Invalid store tome purchase: " + tomeId)
			return
		}

		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
		context.SetData("reward", reward)

	case data.StoreCategoryCards:

	}

	player.Save(context.DB)
}
