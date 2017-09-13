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

    specialOffer := player.Store.SpecialOffer
    cards := player.Store.Cards

    context.SetData("storeItems", player.GetCurrentStoreOffers(context))

    if(specialOffer.ExpirationDate != player.Store.SpecialOffer.ExpirationDate) {
        context.SetData("newSpecialOffer", player.Store.SpecialOffer.Name)
    }
	
	count := 0
	for i := 0; i < len(cards) - 1; i++ {
		if (cards[i].ExpirationDate != player.Store.Cards[i].ExpirationDate) {
			count++
		}
	}
	context.SetData("newCardsCount", count)
}