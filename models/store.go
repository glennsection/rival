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
	if player.PurchaseResetTime < date {
		player.PurchaseResetTime = util.TimeToTicks(time.Now())
		
		for i,_ := range player.Cards {
			player.Cards[i].PurchaseCount = 0
		}

		player.Save(context)
	}

	rand.Seed(player.PurchaseResetTime)

	// get individual card offers
	storeCards := make([]data.StoreData, 0)
	cardTypes := [...]string{"COMMON","COMMON","RARE","EPIC"}

	for _,cardType := range cardTypes {
		id, storeCard := player.GetStoreCard(cardType, storeCards)
		if id == nil || storeCard == nil {
			continue
		}

		storeCards = append(storeCards, *storeCard)
	}

	return storeCards
}

func (player *Player) GetStoreCard(rarity string, storeCards []data.StoreData) (*data.DataId, *data.StoreData) {
	// get cards of the desired rarity
	getCard := func(card *data.CardData) bool {
		for _,item := range storeCards { // ensure no duplicates
			if item.Name == card.Name {
				return false
			}
		}

		// check to see if the card is eligible for purchase
		if cardRef, hasCard := player.HasCard(data.ToDataId(card.Name)); hasCard {
			level := cardRef.GetPotentialLevel()
			if !(data.CanPurchaseCard(rarity, level)) {
				return false
			}
		}

		return card.Rarity == rarity // ensure rarity is correct
	}
	cardIds := data.GetCards(getCard)

	if len(cardIds) == 0 {
		return nil, nil
	}

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

	return &cardId, &storeCard
}

func (player *Player) GetCardCost(id data.DataId) float64 {
	level := 1
	rarity := data.GetCard(id).Rarity
	purchaseCount := 0

	if cardRef, hasCard := player.HasCard(id); hasCard {
		level = cardRef.GetPotentialLevel()
		purchaseCount = cardRef.PurchaseCount
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