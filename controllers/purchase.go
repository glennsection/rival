package controllers

import (
	"bloodtales/system"
)

func HandlePurchase(application *system.Application) {
	application.HandleAPI("/purchase/tome", system.TokenAuthentication, PurchaseTome)
}

func PurchaseTome(context *system.Context) {
	// parse parameters
	tomeId := context.Params.GetRequiredString("tomeId")

	player := context.GetPlayer()

	reward, err := player.ClaimTome(context.DB, tomeId)

	if err != nil {
		panic(err)
	}

	if reward == nil {
		context.Fail("Invalid store purchase")
		return
	}

	context.Data = reward
}
