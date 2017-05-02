package controllers

import (
	"bloodtales/system"
	"bloodtales/data"
)

func HandlePurchase(application *system.Application) {
	application.HandleAPI("/purchase", system.TokenAuthentication, Purchase)
}

func Purchase(context *system.Context) {
	// parse parameters
	itemId := context.Params.GetRequiredString("itemId")

	// get store item
	storeItem := data.GetStoreItem(data.ToDataId(itemId))
	if storeItem == nil {
		context.Fail("Invalid store purchase")
		return
	}

	// get player
	player := context.GetPlayer()

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

	case data.StoreCategoryStandardCurrency:
		player.StandardCurrency += storeItem.Quantity

	case data.StoreCategoryTomes:
		// claim tome
		tomeId := storeItem.ItemID
		reward, err := player.ClaimTome(context.DB, tomeId)
		if err != nil {
			panic(err)
		}
		if reward == nil {
			context.Fail("Invalid store tome purchase: " + tomeId)
			return
		}

		context.Data = reward

	case data.StoreCategoryCards:

	}
}
