package controllers

import (
	"fmt"
	"time"
	"math/rand"
	"encoding/json"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/data"
)

func HandleCard() {
	HandleGameAPI("/card/upgrade", system.TokenAuthentication, UpgradeCard)
	HandleGameAPI("/card/craft", system.TokenAuthentication, CraftCard)
}

func UpgradeCard(context *util.Context) {
	// parse parameters
	id := context.Params.GetRequiredString("cardId")

	player := GetPlayer(context)
	cardIndexes := player.GetMapOfCardIndexes()
	index, valid := cardIndexes[data.ToDataId(id)]
	if !valid {
		context.Fail("Invalid ID")
		return
	}

	card := &player.Cards[index]
	levelData := data.GetCardProgressionData(card.GetData().Rarity, card.Level)

	if player.StandardCurrency < levelData.Cost {
		context.Fail("Insufficient funds")
		return
	} else {
		if card.CardCount < levelData.CardsNeeded  {
			context.Fail("Requirements not met")
			return
		}
	}

	player.StandardCurrency -= levelData.Cost
	player.XP += levelData.XP
	card.CardCount -= levelData.CardsNeeded
	card.Level += 1

	player.SetDirty(models.PlayerDataMask_All)
	player.Save(context.DB)

	context.SetData("card", card)
}

func CraftCard(context *util.Context) {
	// parse parameters
	rarity := context.Params.GetRequiredString("rarity")
	cardsJs := context.Params.GetRequiredString("cards")

	// validate the query
	var cards map[string]int
	json.Unmarshal([]byte(cardsJs), &cards)
	if len(cards) == 0 {
		context.Fail("Malformed Request")
		return
	}

	player := GetPlayer(context)
	cost := data.GetCraftingCost(rarity)
	cardsNeeded := data.GetCraftingXpNeeded(rarity)

	cardsSupplied := 0
	for cardId, num := range cards {
		dataId := data.ToDataId(cardId)
		if card, hasCard := player.HasCard(dataId); hasCard && card.CardCount >= num {
			cardsSupplied += num
			// deduct the cards supplied - if we fail later, we won't update the db and the change wont stick
			card.CardCount -= num
		} else {
				context.Fail(fmt.Sprintf("Insufficient cards of type %s", cardId))
				return
		}
	}

	fmt.Println(fmt.Sprintf("Cards Needed: %d", cardsNeeded))

	// final validation step: make sure the user supplied the correct amount of cards and can afford the exchange
	if cardsSupplied == 0 || cardsSupplied % cardsNeeded != 0 {
		context.Fail("Insufficient Cards")
		return
	} else {
		if player.StandardCurrency < cost {
			context.Fail("Insufficient Funds")
			return
		}
	}

	// subtract the cost of the transaction and add (cardsSupplied/cost) random cards
	player.StandardCurrency -= cost

	numCards := cardsSupplied / cardsNeeded

	accountLevel := player.GetLevel()
	getCards := func(card *data.CardData) bool {
		return card.Rarity == rarity && card.Tier <= accountLevel
	}
	cardSlice := data.GetCards(getCards)

	rand.Seed(time.Now().UnixNano())

	newCards := make([]string, 0)
	for numCards > 0 {
		dataId := cardSlice[rand.Intn(len(cardSlice))]
		newCards = append(newCards, data.ToDataName(dataId))
		player.AddCards(dataId, 1)
		numCards--
	}

	player.Save(context.DB)
	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
	context.SetData("cards", newCards)
}