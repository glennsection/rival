package controllers

import ( 
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleStore() {
	handleGameAPI("/store/offers", system.TokenAuthentication, GetStoreOffers)
}

func GetStoreOffers(context *util.Context) {
	player := GetPlayer(context)
	offers := map[string]interface{}{}

	currentOffers := player.GetCurrentStoreOffers(context)

	//since wyrmtale can't deserialize a dictionary, we need to convert our map into an array
	storeItems := make([]models.StoreItem, 0)
	for _, storeItem := range currentOffers {
		storeItems = append(storeItems, storeItem)
	}

	offers["storeItems"] = storeItems

	context.Data = offers
}