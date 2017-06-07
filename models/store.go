package models

import ( 
	"time"
	"math/rand"
	"sort"
	"fmt"

	"bloodtales/data"
	"bloodtales/util"
)

func (player *Player) GetStoreCards(context *util.Context) []data.StoreData {
	// seed random with current utc date + unique identifer
	year, month, day := time.Now().UTC().Date() 
	date := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

	// ensure our card purchase counts are up to date
	reset := false
	if player.PurchaseResetTime < date {
		player.PurchaseResetTime = util.TimeToTicks(time.Now())
		reset = true
	}

	rand.Seed(player.PurchaseResetTime)

	// get individual card offers
	storeCards := make([]data.StoreData, 0)
	cardTypes := [...]string{"COMMON","COMMON","RARE","EPIC"}

	for _,cardType := range cardTypes {
		id, storeCard := player.GetStoreCard(cardType, storeCards)
		storeCards = append(storeCards, storeCard)

		// reset purchase counts if necessary
		if reset { // should check this condition first before iterating through n cards in HasCard
			if card,ok := player.HasCard(id); ok {
				card.PurchaseCount = 0
			}
		}
	}

	// if we've reset purchase counts, save the changes to the db
	if reset {
		player.Save(context)
	}

	return storeCards
}

func (player *Player) GetStoreCard(rarity string, storeCards []data.StoreData) (data.DataId, data.StoreData) {
	// get cards of the desired rarity
	getCard := func(card *data.CardData) bool {
		for _,item := range storeCards { // ensure no duplicates
			if item.Name == card.Name {
				return false
			}
		}

		return card.Rarity == rarity // ensure rarity is correct
	}
	cardIds := data.GetCards(getCard)

	// sort these cards to ensure we get the same cards for the generated index every time
	sort.Sort(data.DataIdCollection(cardIds))

	// select a card
	cardId := cardIds[rand.Intn(len(cardIds))]
	card := data.GetCard(cardId)

	storeCard := data.StoreData {
		Name: card.Name,
		DisplayName: fmt.Sprintf("%s_NAME", card.Name),
		Image: card.Portrait,
		Category: data.StoreCategoryCards,
		ItemID: card.Name,
		Quantity: 1,
		Currency: data.CurrencyStandard,
		Cost: player.GetCardCost(cardId),
	}

	return cardId, storeCard
}

func (player *Player) GetCardCost(id data.DataId) float64 {
	level := 1
	var purchaseCount int
	var rarity string

	if cardRef, hasCard := player.HasCard(id); hasCard {
		level = cardRef.GetPotentialLevel()
		purchaseCount = cardRef.PurchaseCount
		rarity = data.GetCard(cardRef.DataID).Rarity
	}

	baseCost := data.GetCardCost(rarity, level)

	return float64(baseCost + (purchaseCount * baseCost))
}

func (player *Player) HandleCardPurchase(storeItem *data.StoreData) {
		id := data.ToDataId(storeItem.Name)

		player.AddCards(id, storeItem.Quantity)

		storeItem.Cost = player.GetCardCost(id)

		cardRef,_ := player.HasCard(id)
		cardRef.PurchaseCount++
}