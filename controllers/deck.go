package controllers

import (
	"bloodtales/data"
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/system"
)

func handleDeck() {
	handleGameAPI("/deck/setLeader", system.TokenAuthentication, SetLeaderCard)
	handleGameAPI("/deck/setCard", system.TokenAuthentication, SetDeckCard)
	handleGameAPI("/deck/switch", system.TokenAuthentication, SwitchDeck)
}

func SetLeaderCard(context *util.Context) {
	// parse parameters
	cardId := context.Params.GetRequiredString("cardId")
	cardDataId := data.ToDataId(cardId)

	// get player
	player := GetPlayer(context)

	//validate params
	if card := player.GetCard(cardDataId); card == nil {
		context.Fail("Invalid ID")
		return
	}

	deck := &(player.Decks[player.CurrentDeck])
	deck.SetLeaderCard(cardDataId)

	player.SetDirty(models.PlayerDataMask_Deck)

	player.Save(context)
}

func SetDeckCard(context *util.Context) {
	// parse parameters
	cardId := context.Params.GetRequiredString("cardId")
	deckIndex := context.Params.GetRequiredInt("index")
	cardDataId := data.ToDataId(cardId)

	// get player
	player := GetPlayer(context)

	//validate params
	if deckIndex >= len(player.Decks[player.CurrentDeck].CardIDs) {
		context.Fail("Index out of range")
		return
	}

	if card := player.GetCard(cardDataId); card == nil {
		context.Fail("Invalid ID")
		return
	}

	deck := &(player.Decks[player.CurrentDeck])
	deck.SetDeckCard(cardDataId, deckIndex)

	player.SetDirty(models.PlayerDataMask_Deck)

	player.Save(context)
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

	player.SetDirty(models.PlayerDataMask_Loadout)

	player.Save(context)
}