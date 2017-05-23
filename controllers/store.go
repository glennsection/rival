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

	//TODO get banner data
	//offers["banner"] = banner

	// Get Cards
	cards := player.GetStoreCards()
	for _, card := range cards {
		storeItems = append(storeItems, card);
	}
	offers["storeItems"] = storeItems

	context.Data = offers
}