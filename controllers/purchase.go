package controllers

import (
	"gopkg.in/mgo.v2/bson"
	"time"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/data"
	"bloodtales/util"
)

func handlePurchase() {
	handleGameAPI("/purchase", system.TokenAuthentication, Purchase)
}

func Purchase(context *util.Context) {
	// parse parameters
	itemId := context.Params.GetRequiredString("itemId")

	// get player
	player := GetPlayer(context)

	// get store item
	storeItem := data.GetStoreItem(data.ToDataId(itemId))
	if storeItem == nil {
		// item is not a default store item so check to see if it is a card for sale
		cards := player.GetStoreCards(context)
		for _, card := range cards {
			if itemId == card.Name {
				storeItem = &card
				break
			}
		}

		if storeItem == nil {
			context.Fail("Invalid store purchase")
			return
		}
	}

	// check store item currency cost and store the string converted currency type for tracking
	var currencyType string
	switch storeItem.Currency {

	case data.CurrencyReal:
		currencyType = "USD"
		// TODO - verify mobile store receipt

	case data.CurrencyPremium:
		currencyType = "Premium"
		cost := int(storeItem.Cost)
		if player.PremiumCurrency < cost {
			context.Fail("Insufficient funds")
			return
		}
		player.PremiumCurrency -= cost

	case data.CurrencyStandard:
		currencyType = "Standard"
		cost := int(storeItem.Cost)
		if player.StandardCurrency < cost {
			context.Fail("Insufficient funds")
			return
		}
		player.StandardCurrency -= cost

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
		reward, err := player.ClaimTome(context, tomeId)
		util.Must(err)
		
		if reward == nil {
			context.Fail("Invalid store tome purchase: " + tomeId)
			return
		}

		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
		context.SetData("reward", reward)

	case data.StoreCategoryCards:
		player.HandleCardPurchase(storeItem)
		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
		context.SetData("storeItem", storeItem)
	}

	InsertTracking(context, "purchase", bson.M { "time": util.TimeToTicks(time.Now().UTC()),
													"productId":storeItem.Name,
													"price":storeItem.Cost,
													"currency":currencyType,
													"receipt":"" }, 0)

	player.Save(context)
}
