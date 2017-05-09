package controllers

import (
	"bloodtales/data"
	"bloodtales/system"
)

func HandleDeck(application *system.Application) {
	application.HandleAPI("/deck/setLeader", system.TokenAuthentication, SetLeaderCard)
	application.HandleAPI("/deck/setCard", system.TokenAuthentication, SetDeckCard)
	application.HandleAPI("/deck/switch", system.TokenAuthentication, SwitchDeck)
}

func SetLeaderCard(context *system.Context) {
	// parse parameters
	cardId := context.Params.GetRequiredString("cardId")

	// get player
	player := context.GetPlayer()

	//validate params
	cardIndexes := player.GetMapOfCardIndexes()
	index, valid := cardIndexes[data.ToDataId(cardId)]
	if !valid {
		context.Fail("Invalid ID")
		return
	}

	deck := &(player.Decks[player.CurrentDeck])
	deck.SetLeaderCard(index)
	context.Data = deck

	player.Update(context.DB)
}

func SetDeckCard(context *system.Context) {
	// parse parameters
	cardId := context.Params.GetRequiredString("cardId")
	deckIndex := context.Params.GetRequiredInt("index")

	// get player
	player := context.GetPlayer()

	//validate params
	if deckIndex > len(player.Decks[player.CurrentDeck].CardIDs) {
		context.Fail("Index out of range")
		return
	}

	cardIndexes := player.GetMapOfCardIndexes()
	index, valid := cardIndexes[data.ToDataId(cardId)]
	if !valid {
		context.Fail("Invalid ID")
		return
	}

	deck := &(player.Decks[player.CurrentDeck])
	deck.SetDeckCard(index, deckIndex)
	context.Data = deck

	player.Update(context.DB)
}

func SwitchDeck(context *system.Context) {
	// parse parameters
	currentDeck := context.Params.GetRequiredInt("currentDeck")

	// get player
	player := context.GetPlayer()

	// validate currentDeck
	if currentDeck < 0 || currentDeck > len(player.Decks) {
		context.Fail("Index out of range")
		return
	}

	player.CurrentDeck = currentDeck
	player.Update(context.DB)
}