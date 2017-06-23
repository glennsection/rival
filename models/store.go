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
	currentDate := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
	tomorrow := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1))

	storeCards := make([]data.StoreData, 0)

	// ensure our card purchase counts are up to date
	if player.PurchaseResetTime < currentDate {
		player.PurchaseResetTime = util.TimeToTicks(time.Now())
		player.SpecialOffers = map[string]data.StoreData {}
		
		for i,_ := range player.Cards {
			player.Cards[i].PurchaseCount = 0
		}

		rand.Seed(player.PurchaseResetTime)

		// get individual card offers
		var cardTypes []string
		if player.GetRankTier() == 6 {
			cardTypes = []string{"COMMON","COMMON","RARE","EPIC","LEGENDARY"}
		} else {
			cardTypes = []string{"COMMON","COMMON","RARE","EPIC"}
		}

		for _,cardType := range cardTypes {
			_,storeCard := player.GetStoreCard(cardType, storeCards)

			if storeCard != nil {
				storeCard.ExpirationDate = tomorrow
				player.SpecialOffers[storeCard.Name] = *storeCard
				storeCards = append(storeCards, *storeCard)
			}
		}

		player.Save(context)
	}

	if len(storeCards) == 0 { //entering this block means the users card selection has not been reset in this call
		for _, specialOffer := range player.SpecialOffers {
			if specialOffer.Category == data.StoreCategoryCards {
				storeCards = append(storeCards, specialOffer)
			}
		}
	}
	

	return storeCards
}

func (player *Player) GetStoreCard(rarity string, storeCards []data.StoreData) (*data.DataId, *data.StoreData) {
	// get cards of the desired rarity
	getCard := func(card *data.CardData) bool {
		for _,item := range storeCards { // ensure no duplicates
			if item.Name == card.Name || card.Tier > player.GetLevel() {
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
		ItemID: card.Name,
		Category: data.StoreCategoryCards,
		Currency: data.CurrencyStandard,
		Cost: player.GetCardCost(cardId),
		Availability: data.Availability_Limited,
		IsOneTimeOffer: false,
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

	player.AddCards(id, 1)

	cardRef,_ := player.HasCard(id)
	cardRef.PurchaseCount++
	
	if offer, exists := player.SpecialOffers[storeItem.Name]; exists {
		offer.Cost = player.GetCardCost(id)
	}

	storeItem.Cost = player.GetCardCost(id)
}