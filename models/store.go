package models

import (
	"encoding/json"
	"math/rand"
	"math"
	"sort"
	"strconv"
	"time"

	"bloodtales/data"
	"bloodtales/util"
)

type OfferHistory struct {
	ExpirationDate 				time.Time
	Purchased 					bool
}

type StoreHistory struct {
	LastUpdate 					int64 						`bson:"lu"`
	Cards 	 					[]StoreItem 				`bson:"co"`
	SpecialOffer 				StoreItem 					`bson:"so"`
	SpecialOfferQueue 			OfferQueue 					`bson:"oq"`
	SpecialOfferHistory 		map[string]OfferHistory		`bson:"ed"`
}

type StoreItem struct {
	Name 						string

	ItemID    					string
	Category  					data.StoreCategory
	RewardIDs 					[]data.DataId

	Currency 					data.CurrencyType
	Cost     					float64

	NumAvailable 				int 
	BulkCost 					float64

	ExpirationDate 				int64
}

// custom marshalling
func (storeItem *StoreItem) MarshalJSON() ([]byte, error) {
	client := map[string]interface{}{}

	client["id"] = storeItem.Name
	client["itemId"] = storeItem.ItemID
	client["cost"] = storeItem.Cost
	client["numAvailable"] = storeItem.NumAvailable
	client["bulkCost"] = storeItem.BulkCost

	var err error
	err = nil

	if client["category"], err = data.StoreCategoryToString(storeItem.Category); err != nil {
		return nil, err
	}

	if client["currency"], err = data.CurrencyTypeToString(storeItem.Currency); err != nil {
		return nil, err
	}

	clientRewards := make([]string, 0)
	for _, reward := range storeItem.RewardIDs {
		clientRewards = append(clientRewards, data.ToDataName(reward))
	}
	client["rewardIds"] = util.StringArrayToString(clientRewards)

	if storeItem.ExpirationDate > 0 {
		client["expirationDate"] = strconv.FormatInt(storeItem.ExpirationDate-util.TimeToTicks(time.Now().UTC()), 10)
	}

	return json.Marshal(client)
}

func (player *Player) InitStore() {
	defaultSpecialOffer := StoreItem { 
		Name: "", 
		ExpirationDate: 0, 
	}

	player.Store = StoreHistory {
		LastUpdate: 0,
		SpecialOffer: defaultSpecialOffer,
		SpecialOfferHistory: map[string]OfferHistory {},
	}
}

func (player *Player) RecordSpecialOfferPurchase() {
	id := player.Store.SpecialOffer.Name

	offerHistory := player.Store.SpecialOfferHistory[id]
	offerHistory.Purchased = true
	player.Store.SpecialOfferHistory[id] = offerHistory

	player.Store.SpecialOffer.ExpirationDate = 0
}

func (player *Player) GetCurrentStoreOffers(context *util.Context) []StoreItem {
	currentOffers := make([]StoreItem, 0)

	year, month, day := time.Now().UTC().Date() 
	currentDate := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

	// get a special offer if one is currently available to the player
	currentSpecialOffer := player.getSpecialOffer(currentDate)
	if currentSpecialOffer != nil {
		currentOffers = append(currentOffers, *currentSpecialOffer)
	}

	// next check to see if we need to generate new card offers
	if currentDate > player.Store.LastUpdate || len(player.Store.Cards) == 0 {
		player.getStoreCards()
	}

	// add the current day's card offers to the slice
	for _, storeCard := range player.Store.Cards {
		currentOffers = append(currentOffers, storeCard)
	}

	// Next retrieve the rest of the store's currently available offerings
	costMultiplier := data.GetLeagueData(data.GetLeague(data.GetRank(player.RankPoints).Level)).TomeCostMultiplier

	storeItems := data.GetStoreItemDataCollection()
	for _, storeItemData := range storeItems {
		if !player.canPurchase(storeItemData, currentDate) {
			continue
		}

		cost := storeItemData.Cost
		if storeItemData.Category == data.StoreCategoryTomes { cost = math.Floor(cost * costMultiplier) }

		currentOffers = append(currentOffers, StoreItem {
			Name: storeItemData.Name,
			ItemID: storeItemData.ItemID,
			Category: storeItemData.Category,
			RewardIDs: storeItemData.RewardIDs,
			Currency: storeItemData.Currency,
			Cost: cost,
			NumAvailable: 1,
			BulkCost: storeItemData.Cost,
		})
	}

	player.Store.LastUpdate = currentDate
	player.Save(context)

	return currentOffers
}

func (player *Player) canPurchase(storeItemData *data.StoreItemData, currentDate int64) bool {
	// first confirm the offer is available in the player's current league
	if storeItemData.League != 0 && storeItemData.League != player.GetRankTier() {
		return false
	}

	// next confirm the offer is available at this time
	if (storeItemData.AvailableDate > 0 && storeItemData.AvailableDate > currentDate) || 
	   (storeItemData.ExpirationDate > 0 && currentDate > storeItemData.ExpirationDate) {
		return false
	}

	if storeItemData.Category == data.StoreCategorySpecialOffers {
		// check to see if the user has ever purchased this item before or if it already exists in their queue
		if _, hasEntry := player.Store.SpecialOfferHistory[storeItemData.Name]; hasEntry || player.Store.SpecialOfferQueue.Contains(data.ToDataId(storeItemData.Name)) {
			return false
		}

		//TODO check cooldowns
	}

	return true
}

func (player *Player) getSpecialOffer (currentDate int64) *StoreItem {
	// first check to see if the current special offer is still valid
	if player.Store.SpecialOffer.ExpirationDate > util.TimeToTicks(time.Now().UTC()) {
		return &player.Store.SpecialOffer
	} 

	if currentDate > player.Store.LastUpdate {
		// Populate our special offer queue
		specialOffers := data.GetSpecialOfferCollection()
		for _, specialOfferData := range specialOffers {
			if player.canPurchase(specialOfferData, currentDate) {
				player.Store.SpecialOfferQueue.Push(data.ToDataId(specialOfferData.Name))
			}
		}

		//if we have any available offers in our queue, pop the highest priority one
		if !player.Store.SpecialOfferQueue.IsEmpty() {
			specialOfferData := data.GetStoreItemData(player.Store.SpecialOfferQueue.Pop())
			expirationDate := time.Now().UTC().AddDate(0, 0, specialOfferData.Duration)

			//now create a StoreItem and assign it to the current special offer field
			player.Store.SpecialOffer = StoreItem {
				Name: specialOfferData.Name,
				ItemID: specialOfferData.ItemID,
				Category: specialOfferData.Category,
				RewardIDs: specialOfferData.RewardIDs,
				Currency: specialOfferData.Currency,
				Cost: specialOfferData.Cost,
				NumAvailable: 1,
				BulkCost: specialOfferData.Cost,
				ExpirationDate: util.TimeToTicks(expirationDate),
			}

			//create an OfferHistory record for the offer
			player.Store.SpecialOfferHistory[specialOfferData.Name] = OfferHistory {
				ExpirationDate: expirationDate,
				Purchased: false,
			}

			return &player.Store.SpecialOffer
		}
	}

	return nil
}

func (player *Player) getStoreCards() {

	for i := range player.Cards {
		player.Cards[i].PurchaseCount = 0
	}

	player.Store.Cards = make([]StoreItem, 0)

	rand.Seed(time.Now().UnixNano())

	// get individual card offers
	var cardTypes []string
	if player.GetRankTier() == 6 {
		cardTypes = []string{"COMMON", "COMMON", "RARE", "EPIC", "LEGENDARY"}
	} else {
		cardTypes = []string{"COMMON", "COMMON", "RARE", "EPIC"}
	}

	year, month, day := time.Now().UTC().Date()
	expDate := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1))

	for _, cardType := range cardTypes {

		if storeItem := player.getStoreCard(cardType, expDate); storeItem != nil {
			player.Store.Cards = append(player.Store.Cards, *storeItem)
		}
	}
}

func (player *Player) getStoreCard(rarity string, expirationDate int64) *StoreItem {
	// get cards of the desired rarity
	cardIds := data.GetCards( func(card *data.CardData) bool {
 		for _, storeCard := range player.Store.Cards { // ensure no duplicates
 			if storeCard.Name == card.Name {
 				return false
 			}
		}

		if card.Tier > player.GetLevel() { // ensure player only gets cards they can earn in tomes
			return false
		}

		return card.Rarity == rarity // ensure rarity is correct
	})

	if len(cardIds) == 0 {
		return nil
	}

	// sort these cards to ensure we get the same cards for the generated index every time
	sort.Sort(data.DataIdCollection(cardIds))

	// select a card
	cardId := cardIds[rand.Intn(len(cardIds))]
	card := data.GetCard(cardId)

	storeCard := &StoreItem {
		Name: card.Name,
		ItemID: card.Name,
		Category: data.StoreCategoryCards,
		Currency: data.CurrencyStandard,
		Cost: getCardCost(rarity, 0),
		NumAvailable: data.GetMaxPurchaseCount(rarity),
		BulkCost: getBulkCost(rarity, 0),
		ExpirationDate: expirationDate,
	}

	return storeCard
}

func getCardCost(rarity string, purchaseCount int) float64 {
	baseCost := data.GetCardCost(rarity)

	return float64(baseCost + (purchaseCount * baseCost))
}

func getBulkCost(rarity string, purchaseCount int) float64 {
	baseCost := data.GetCardCost(rarity)
	maxPurchaseCount := data.GetMaxPurchaseCount(rarity)

	bulkCost := 0

	for i := purchaseCount; i <= maxPurchaseCount; i++ {
		bulkCost += baseCost + (i * baseCost)
	}

	return float64(bulkCost)
}

func (player *Player) HandleCardPurchase(storeItem *StoreItem, bulk bool) {
	name := storeItem.Name
	id := data.ToDataId(name)
	rarity := data.GetCard(id).Rarity
	var numPurchased int

	if bulk { // player has purchased all remaining cards
		numPurchased = storeItem.NumAvailable
		storeItem.NumAvailable = 0
	} else { // single card purchase
		numPurchased = 1
		storeItem.NumAvailable -= 1

		purchaseCount := data.GetMaxPurchaseCount(rarity) - storeItem.NumAvailable
		storeItem.BulkCost -= storeItem.Cost
		storeItem.Cost = getCardCost(rarity, purchaseCount)
	}

	player.AddCards(id, numPurchased)

	for i := range player.Store.Cards {
		if storeItem.Name == player.Store.Cards[i].Name {
			player.Store.Cards[i] = *storeItem
			break
		}
	}
}
