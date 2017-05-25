package controllers

import ( 
	"bloodtales/data"
	"bloodtales/system"
	"bloodtales/util"
)

func HandleStore() {
	HandleGameAPI("/store/offers", system.TokenAuthentication, GetStoreOffers)
}

func GetStoreOffers(context *util.Context) {
	player := GetPlayer(context)
	offers := map[string]interface{}{}
	storeItems := data.GetStoreItems()

	// Get Banner

	// Get Special Offers

	// Get Cards
	cards := player.GetStoreCards(context.DB)
	for _, card := range cards {
		storeItems = append(storeItems, card);
	}
	offers["storeItems"] = storeItems

	context.Data = offers
}