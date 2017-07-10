package controllers

import (
	"gopkg.in/mgo.v2/bson"
	"time"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/data"
	"bloodtales/util"
	"bloodtales/log"
)

func handlePurchase() {
	handleGameAPI("/purchase", system.TokenAuthentication, Purchase)
}

func Purchase(context *util.Context) {
	// parse parameters
	id := context.Params.GetRequiredString("id")

	// get player
	player := GetPlayer(context)

	// get current offers
	currentOffers := player.GetCurrentStoreOffers(context)


	storeItem, valid := currentOffers[id]
	if !valid {
		context.Fail("Invalid store purchase")
		return
	}

	// check if this is a one time purchase item
	if storeItem.IsOneTimeOffer {
		if _, hasEntry := player.Store.Purchases[storeItem.Name]; hasEntry {
			context.Fail("Item is a one-time offer. Cannot purchase again.")
			return
		}
	}

	// check store item currency cost and store the string converted currency type for tracking
	var currencyType string
	purchasePrice := storeItem.Cost // need to cache this for analytics, since some store items change price after purchase

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

	// add rewards
	if storeItem.Category != data.StoreCategoryCards {
		rewards := player.GetRewards(storeItem.RewardIDs)

		for _, reward := range rewards {
			player.AddRewards(reward, nil)
		}
		context.SetData("rewards", rewards)
	}

	// handle store item category
	switch storeItem.Category {

	case data.StoreCategoryPremiumCurrency:
		player.SetDirty(models.PlayerDataMask_Currency)

	case data.StoreCategoryStandardCurrency:
		player.SetDirty(models.PlayerDataMask_Currency)

	case data.StoreCategoryTomes:
		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
	
		// analytics
		tome := data.GetTome(data.ToDataId(storeItem.ItemID))
		if tome != nil {
			InsertTracking(context, "tomeOpened", bson.M { "rarity": tome.Rarity }, 0)
		} else {
			log.Errorf("Failed find data for purchased tome: %s", storeItem.ItemID)
		}
		
	case data.StoreCategoryCards:
		player.HandleCardPurchase(&storeItem)
		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
		context.SetData("storeItem", &storeItem) //include the updated store item
	}

	player.RecordPurchase(storeItem.Name)

	InsertTracking(context, "purchase", bson.M { "time": util.TimeToTicks(time.Now().UTC()),
													"productId":storeItem.Name,
													"price":purchasePrice,
													"currency":currencyType,
													"receipt":"" }, 0)

	player.Save(context)
}
