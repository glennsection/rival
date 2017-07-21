package controllers

import ( 
	"bloodtales/system"
	"bloodtales/util"
)

func handleStore() {
	handleGameAPI("/store/offers", system.TokenAuthentication, GetStoreOffers)
}

func GetStoreOffers(context *util.Context) {
	player := GetPlayer(context)
	offers := map[string]interface{}{}

	offers["storeItems"] = player.GetCurrentStoreOffers(context)

	context.Data = offers
}