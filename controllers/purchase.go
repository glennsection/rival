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
	bulk := context.Params.GetBool("bulk", false)

	// get player
	player := GetPlayer(context)

	// get current offers
	currentOffers := player.GetCurrentStoreOffers(context)

	var storeItem *models.StoreItem

	valid := false
	for _, item := range currentOffers {
		valid = item.Name == id
		if valid {
			storeItem = &item
			break
		}
	}

	if !valid || storeItem.NumAvailable == 0 {
		context.Fail("Invalid store purchase")
		return
	}

	// if this is a special offer, ensure the player has not already purchased it
	if storeItem.Category == data.StoreCategorySpecialOffers {
		if offerHistory, hasEntry := player.Store.SpecialOfferHistory[storeItem.Name]; hasEntry && offerHistory.Purchased {
			context.Fail("Item is a one-time offer. Cannot purchase again.")
			return
		}
	}

	// check store item currency cost and store the string converted currency type for tracking
	var currencyType string
	purchasePrice := storeItem.Cost // need to cache this for analytics, since some store items change price after purchase
	currentTime := util.TimeToTicks(time.Now().UTC())

	switch storeItem.Currency {

	case data.CurrencyReal:
		currencyType = "USD"
		// TODO - verify mobile store receipt

	case data.CurrencyPremium:
		currencyType = "Premium"

		var cost int
		if bulk { cost = int(storeItem.BulkCost) } else { cost = int(storeItem.Cost) }

		if player.PremiumCurrency < cost {
			context.Fail("Insufficient funds")
			return
		}
		player.PremiumCurrency -= cost

	case data.CurrencyStandard:
		currencyType = "Standard"

		var cost int
		if bulk { cost = int(storeItem.BulkCost) } else { cost = int(storeItem.Cost) }

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

			if util.HasSQLDatabase() {
				TrackRewardsSQL(context, reward, currentTime)
			}else{
				TrackRewards(context, reward)
			}
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
			if util.HasSQLDatabase() {
				InsertTrackingSQL(context, "tomeOpened", currentTime, tome.Name, "Premium", 1, 0, nil)
			}else{
				InsertTracking(context, "tomeOpened", bson.M { "tomeId": tome.Name }, 0)
			}
		} else {
			log.Errorf("Failed find data for purchased tome: %s", storeItem.ItemID)
		}
		
	case data.StoreCategoryCards:
		player.HandleCardPurchase(storeItem, bulk)
		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
		context.SetData("storeItem", storeItem) //include the updated store item

	case data.StoreCategorySpecialOffers:
		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
		player.RecordSpecialOfferPurchase()
	}


	if util.HasSQLDatabase() {
		InsertTrackingSQL(context, "purchase", currentTime, storeItem.Name, currencyType, 1, purchasePrice, bson.M { "time": currentTime,
													"productId":storeItem.Name,
													"price":purchasePrice,
													"currency":currencyType,
													"receipt":"" })
	} else {
		InsertTracking(context, "purchase", bson.M { "time": currentTime,
													"productId":storeItem.Name,
													"price":purchasePrice,
													"currency":currencyType,
													"receipt":"" }, 0)
	}

	player.Save(context)
}
