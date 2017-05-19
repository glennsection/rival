package controllers

import (
	"bloodtales/data"
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/system"
)

func HandleDeck() {
	HandleGameAPI("/deck/setLeader", system.TokenAuthentication, SetLeaderCard)
	HandleGameAPI("/deck/setCard", system.TokenAuthentication, SetDeckCard)
	HandleGameAPI("/deck/switch", system.TokenAuthentication, SwitchDeck)
}

func SetLeaderCard(context *util.Context) {
	// parse parameters
	cardId := context.Params.GetRequiredString("cardId")
	cardDataId := data.ToDataId(cardId)

	// get player
	player := GetPlayer(context)

	//validate params
	cardIndexes := player.GetMapOfCardIndexes()
	_, valid := cardIndexes[cardDataId]
	if !valid {
		context.Fail("Invalid ID")
		return
	}

	deck := &(player.Decks[player.CurrentDeck])
	deck.SetLeaderCard(cardDataId)

	context.SetDirty([]int64{models.UpdateMask_Deck})

	player.Save(context.DB)
}

func SetDeckCard(context *util.Context) {
	// parse parameters
	cardId := context.Params.GetRequiredString("cardId")
	deckIndex := context.Params.GetRequiredInt("index")

	cardDataId := data.ToDataId(cardId)

	// get player
	player := GetPlayer(context)

	//validate params
	if deckIndex > len(player.Decks[player.CurrentDeck].CardIDs) {
		context.Fail("Index out of range")
		return
	}

	cardIndexes := player.GetMapOfCardIndexes()
	_, valid := cardIndexes[cardDataId]
	if !valid {
		context.Fail("Invalid ID")
		return
	}

	deck := &(player.Decks[player.CurrentDeck])
	deck.SetDeckCard(cardDataId, deckIndex)

	context.SetDirty([]int64{models.UpdateMask_Deck})

	player.Save(context.DB)
}

func SwitchDeck(context *util.Context) {
	// parse parameters
	currentDeck := context.Params.GetRequiredInt("currentDeck")

	// get player
	player := GetPlayer(context)

	// validate currentDeck
	if currentDeck < 0 || currentDeck > len(player.Decks) {
		context.Fail("Index out of range")
		return
	}

	player.CurrentDeck = currentDeck

	context.SetDirty([]int64{models.UpdateMask_Loadout})

	player.Save(context.DB)
}