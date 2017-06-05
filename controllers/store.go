package controllers

import ( 
	"bloodtales/data"
	"bloodtales/system"
	"bloodtales/util"
)

func handleStore() {
	handleGameAPI("/store/offers", system.TokenAuthentication, GetStoreOffers)
}

func GetStoreOffers(context *util.Context) {
	player := GetPlayer(context)
	offers := map[string]interface{}{}
	storeItems := data.GetStoreItems()

	// Get Banner

	// Get Special Offers

	// Get Cards
	cards := player.GetStoreCards(context)
	for _, card := range cards {
		storeItems = append(storeItems, card);
	}
	offers["storeItems"] = storeItems

	context.Data = offers
}