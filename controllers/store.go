package controllers

import ( 
	"time"
	"math/rand"
	"encoding/hex"
	"sort"
	"fmt"

	"bloodtales/data"
	"bloodtales/system"
	"bloodtales/util"
)

func HandleStore() {
	HandleGameAPI("/store/offers", system.TokenAuthentication, GetSpecialOffers)
}

func GetSpecialOffers(context *util.Context) {
	player := GetPlayer(context)

	offers := map[string]interface{}{}

	//TODO get banner data

	//TODO get sale data

	// Get Cards
	//seed random with current utc date + unique identifer
	year, month, day := time.Now().UTC().Date() 
	hexId := player.ID.Hex()
	dst := make([]byte, hex.DecodedLen(len(hexId)))
	uniqueId, _ := hex.Decode(dst, []byte(hexId))
	rand.Seed(data.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)) + int64(uniqueId))

	cards := GetStoreCards()

	offers["cards"] = cards

	context.SetData("offers", offers)
}

func GetStoreCards() []data.StoreData {
	cards := make([]data.StoreData, 0)
	cards = append(cards, GetStoreCard("COMMON"))
	cards = append(cards, GetStoreCard("RARE"))
	cards = append(cards, GetStoreCard("EPIC"))
	return cards
}

func GetStoreCard(rarity string) data.StoreData {
	//get cards of the desired rarity
	getCard := func(card *data.CardData) bool {
		return card.Rarity == rarity
	}
	cards := data.GetCards(getCard)

	//since cards in a map are returned in random order, we need to sort these cards to ensure we get the same cards for the generated index every time
	sort.Sort(data.DataIdCollection(cards))

	card := data.StoreData {
		Name: fmt.Sprintf("STORE_CARD_%s", rarity),
		Image: "",
		Category: data.StoreCategoryCards,
		ItemID: data.GetCard(cards[rand.Intn(len(cards))]).Name,
		Quantity: 1,
		Currency: data.CurrencyPremium,
		Cost: 1, // TODO get cost from Larry's excel sheet
	}

	return card
}