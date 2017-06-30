package models

import ( 
	"time"
	"math/rand"
	"encoding/json"
	"strconv"
	"sort"
	"fmt"

	"bloodtales/data"
	"bloodtales/util"
)

type StoreHistory struct {
	LastUpdate 					int64 						`bson:"lu"`
	CurrentOffers 				map[string]StoreItem 		`bson:"co"`
	Purchases 					map[string][]time.Time 		`bson:"ph"`
	CustomExpirationDates		map[string]int64 			`bson:"ed"`
}

type StoreItem struct {
	Name                    string
	DisplayName 			string
	Description 			string

	ItemID                  string
	Category                data.StoreCategory
	RewardIDs				[]data.DataId

	Currency                data.CurrencyType
	Cost                    float64

	Availability 			data.AvailabilityType
	IsOneTimeOffer 			bool
	ExpirationDate 			int64
}

// custom marshalling
func (storeItem *StoreItem) MarshalJSON() ([]byte, error) {
	client := map[string]interface{} {}

	client["id"] = storeItem.Name
	client["displayName"] = storeItem.DisplayName
	client["descrition"] = storeItem.Description
	client["itemId"] = storeItem.ItemID
	client["cost"] = storeItem.Cost
	client["isOneTimeOffer"] = storeItem.IsOneTimeOffer

	var err error
	err = nil

	if client["category"], err = data.StoreCategoryToString(storeItem.Category); err != nil {
		return nil, err
	}

	if client["currency"], err = data.CurrencyTypeToString(storeItem.Currency); err != nil {
		return nil, err
	}

	if client["availability"], err = data.AvailabilityTypeToString(storeItem.Availability); err != nil {
		return nil, err
	}

	clientRewards := make([]string, 0)
	for _, reward := range storeItem.RewardIDs {
		clientRewards = append(clientRewards, data.ToDataName(reward))
	}
	client["rewardIds"] = util.StringArrayToString(clientRewards)

	if(storeItem.ExpirationDate > 0) {
		client["expirationDate"] = strconv.FormatInt(storeItem.ExpirationDate - util.TimeToTicks(time.Now().UTC()), 10)	
	}

	return json.Marshal(client)
}

func (player *Player) InitStore() {
	player.Store = StoreHistory{
		LastUpdate: 0,
		Purchases: map[string][]time.Time {},
		CustomExpirationDates: map[string]int64 {},
	}
}

func (player *Player) RecordPurchase(id string) {
	if _, hasEntry := player.Store.Purchases[id]; !hasEntry {
		player.Store.Purchases[id] = make([]time.Time, 0)
	}

	player.Store.Purchases[id] = append(player.Store.Purchases[id], time.Now())

	if item, hasOffer := player.Store.CurrentOffers[id]; hasOffer && item.IsOneTimeOffer {
		delete(player.Store.CurrentOffers, id)
	}
}

func (player *Player) GetCurrentStoreOffers(context *util.Context) map[string]StoreItem {
	year, month, day := time.Now().UTC().Date() 
	currentDate := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

	save := false

	if currentDate > player.Store.LastUpdate || len(player.Store.CurrentOffers) == 0 {
		player.Store.LastUpdate = currentDate
		player.Store.CurrentOffers = map[string]StoreItem {}
		player.getStoreCards()

		storeItems := data.GetStoreItemDataCollection()

		for _, storeItemData := range storeItems {

			if !player.canPurchase(storeItemData, currentDate) {
				continue
			}

			if storeItem := player.generateStoreItem(storeItemData); storeItem != nil {
				player.Store.CurrentOffers[(*storeItem).Name] =  *storeItem
			}
		}

		save = true
	}

	if save {
		player.Save(context)
	}

	return player.Store.CurrentOffers
}

func (player *Player) canPurchase(storeItemData *data.StoreItemData, currentDate int64) bool { 
	// first confirm the offer is available in the player's current league
	if storeItemData.League != 0 && storeItemData.League != player.GetRankTier() {
		return false
	}

	// next confirm the offer is available at this time
	if storeItemData.Availability != data.Availability_Permanent {

	  	if (storeItemData.AvailableDate > 0 && storeItemData.AvailableDate > currentDate) || 
	       (storeItemData.ExpirationDate > 0 && currentDate > storeItemData.ExpirationDate) {
			return false
		}
	}

	if storeItemData.IsOneTimeOffer {
		// first check to see if the user has ever purchased this item before
		if _, hasEntry := player.Store.Purchases[storeItemData.Name]; hasEntry {
			return false
		}

		//next check to see if this offer requires a custom expiration date
		if storeItemData.Duration > 0 {
			// first check if we've already generated a date
			if customExpDate, hasDate := player.Store.CustomExpirationDates[storeItemData.Name]; hasDate {
				if currentDate > customExpDate {
					return false
				}
			}
		}
	}

	return true
}

func (player *Player) generateStoreItem(storeItemData *data.StoreItemData) (*StoreItem) {
	storeItem := &StoreItem {
		Name: storeItemData.Name,
		DisplayName: storeItemData.DisplayName,
		Description: storeItemData.Description,
		ItemID: storeItemData.ItemID,
		Category: storeItemData.Category,
		RewardIDs: storeItemData.RewardIDs,
		Currency: storeItemData.Currency,
		Cost: storeItemData.Cost,
		Availability: storeItemData.Availability,
		IsOneTimeOffer: storeItemData.IsOneTimeOffer,
	}

	// client expiration date
	var expirationDate int64

	if storeItemData.IsOneTimeOffer && storeItemData.Duration > 0 {
		// first check if we've already generated a date
		if customExpDate, hasDate := player.Store.CustomExpirationDates[storeItemData.Name]; hasDate {
			expirationDate = customExpDate
		} else {
			// generate a date and store it
			year, month, day := time.Now().UTC().AddDate(0, 0, storeItemData.Duration).Date()
			expirationDate := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))

			if storeItemData.ExpirationDate > 0 && storeItemData.ExpirationDate < expirationDate {
				expirationDate = storeItemData.ExpirationDate
			}

			player.Store.CustomExpirationDates[storeItemData.Name] = expirationDate
		}
	} else {
		expirationDate = storeItemData.ExpirationDate
	}

	storeItem.ExpirationDate = expirationDate

	return storeItem
}

func (player *Player) getStoreCards() {
		
	for i,_ := range player.Cards {
		player.Cards[i].PurchaseCount = 0
	}

	rand.Seed(time.Now().UnixNano())

	// get individual card offers
	var cardTypes []string
	if player.GetRankTier() == 6 {
		cardTypes = []string{"COMMON","COMMON","RARE","EPIC","LEGENDARY"}
	} else {
		cardTypes = []string{"COMMON","COMMON","RARE","EPIC"}
	}

	year, month, day := time.Now().UTC().Date() 
	expDate := util.TimeToTicks(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1))

	for _,cardType := range cardTypes {

		if storeItem := player.getStoreCard(cardType, expDate); storeItem != nil {
			player.Store.CurrentOffers[(*storeItem).Name] =  *storeItem
		}
	}
}

func (player *Player) getStoreCard(rarity string, expirationDate int64) (*StoreItem) {
	// get cards of the desired rarity
	getCard := func(card *data.CardData) bool {

		if _, hasOffer := player.Store.CurrentOffers[card.Name]; hasOffer { // ensure no duplicates
			return false
		}

		if card.Tier > player.GetLevel() { // ensure player only gets cards they can earn in tomes
			return false
		}

		return card.Rarity == rarity // ensure rarity is correct
	}
	cardIds := data.GetCards(getCard)

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
		DisplayName: fmt.Sprintf("%s_NAME", card.Name),
		ItemID: card.Name,
		Category: data.StoreCategoryCards,
		Currency: data.CurrencyStandard,
		Cost: player.getCardCost(cardId),
		Availability: data.Availability_Limited,
		IsOneTimeOffer: false,
		ExpirationDate: expirationDate,
	}

	return storeCard
}

func (player *Player) getCardCost(id data.DataId) float64 {
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

func (player *Player) HandleCardPurchase(storeItem *StoreItem) {
	name := storeItem.Name
	id := data.ToDataId(name)

	player.AddCards(id, 1)

	cardRef,_ := player.HasCard(id)
	cardRef.PurchaseCount++
	
	storeItem.Cost = player.getCardCost(id)
	player.Store.CurrentOffers[storeItem.Name] = *storeItem
}