package models

import ( 
	"time"
	"math/rand"
	"sort"
	"fmt"

	"gopkg.in/mgo.v2"

	"bloodtales/data"
	"bloodtales/util"
)

func (player *Player) GetNumCardsPurchased(rarity string) *int {
	switch rarity {
		case "COMMON":
			return &player.CardsPurchased[0]
		case "RARE":
			return &player.CardsPurchased[1]
		case "EPIC":
			return &player.CardsPurchased[2]
	}

	return nil
}

func (player *Player) GetStoreCards(database *mgo.Database) []data.StoreData {
	// seed random with current utc date + unique identifer
	year, month, day := time.Now().UTC().Date() 
	date := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

	// ensure our card purchase counts are up to date
	if player.PurchaseResetTime < date {
		player.PurchaseResetTime = util.TimeToTicks(time.Now())

		for i, _ := range player.CardsPurchased {
			player.CardsPurchased[i] = 0
		}

		player.Save(database)
	}

	rand.Seed(player.PurchaseResetTime)

	// get individual card offers
	cards := make([]data.StoreData, 0)
	cards = append(cards, player.GetStoreCard("COMMON"))
	cards = append(cards, player.GetStoreCard("RARE"))
	cards = append(cards, player.GetStoreCard("EPIC"))
	return cards
}

func (player *Player) GetStoreCard(rarity string) data.StoreData {
	// get cards of the desired rarity
	getCard := func(card *data.CardData) bool {
		return card.Rarity == rarity
	}
	cards := data.GetCards(getCard)

	// sort these cards to ensure we get the same cards for the generated index every time
	sort.Sort(data.DataIdCollection(cards))

	// select a card
	id := cards[rand.Intn(len(cards))]
	card := data.GetCard(id)

	storeCard := data.StoreData {
		Name: card.Name,
		DisplayName: card.Name,
		Image: card.Portrait,
		Category: data.StoreCategoryCards,
		ItemID: fmt.Sprintf("STORE_CARD_%s", rarity),
		Quantity: 1,
		Currency: data.CurrencyPremium,
		Cost: player.GetCardCost(id, rarity),
	}

	return storeCard
}

func (player *Player) GetCardCost(id data.DataId, rarity string) float64 {
	//TODO need to handle card levels > level 9

	level := 1
	if cardRef, hasCard := player.HasCard(id); hasCard {
		level = cardRef.GetPotentialLevel()
	}
	baseCost := data.GetCardCost(rarity, level)

	return float64(baseCost + (*player.GetNumCardsPurchased(rarity) * baseCost))
}

func (player *Player) HandleCardPurchase(storeItem *data.StoreData) {
		id := data.ToDataId(storeItem.Name)
		var rarity string

		fmt.Println(fmt.Sprintf("Name: %s", storeItem.Name))

		switch storeItem.ItemID {

		case "STORE_CARD_COMMON":
			rarity = "COMMON"

		case "STORE_CARD_RARE":
			rarity = "RARE"

		case "STORE_CARD_EPIC":
			rarity = "EPIC"
		}

		(*player.GetNumCardsPurchased(rarity))++
		storeItem.Cost = player.GetCardCost(id, rarity)

		player.AddCards(id, storeItem.Quantity)
}