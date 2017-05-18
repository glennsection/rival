package controllers

import ( 
	"time"
	"math/rand"
	"encoding/hex"

	"bloodtales/data"
	"bloodtales/system"
)

func HandleStore(application *system.Application) {
	HandleGameAPI(application, "/store/offers", system.TokenAuthentication, GetSpecialOffers)
}

func GetSpecialOffers(context *system.Context) {
	player := GetPlayer(context)

	//seed random with current utc date + unique identifer
	year, month, day := time.Now().UTC().Date() 
	uniqueId, _ := hex.Decode([]byte{}, []byte(player.ID.Hex()))
	rand.Seed(data.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)) + int64(uniqueId))

	cards := make([]*data.CardData, 0)
	cards = append(cards, GetStoreCard("COMMON"))
	cards = append(cards, GetStoreCard("RARE"))
	cards = append(cards, GetStoreCard("EPIC"))

	context.Data = cards
}

func GetStoreCard(rarity string) *data.CardData {
	getCard := func(card *data.CardData) bool {
		return card.Rarity == rarity
	}
	cards := data.GetCards(getCard)
	card := data.GetCard(cards[rand.Intn(len(cards))])
	return card
}