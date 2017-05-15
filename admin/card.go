package admin

import (
	"fmt"
	
	// "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminCards(application *system.Application) {
	handleAdminTemplate(application, "/admin/cards/edit", system.TokenAuthentication, EditCard, "")
	handleAdminTemplate(application, "/admin/cards/delete", system.TokenAuthentication, DeleteCard, "")
}

func EditCard(context *system.Context) {
	// parse parameters
	playerId := context.Params.GetRequiredID("playerId")
	cardId := data.DataId(context.Params.GetRequiredInt("card"))

	player, err := models.GetPlayerById(context.DB, playerId)
	util.Must(err)

	for i, card := range player.Cards {
		if card.DataID == cardId {
			player.Cards[i].Level = context.Params.GetRequiredInt("level")
			player.Cards[i].CardCount = context.Params.GetRequiredInt("cardCount")
			player.Cards[i].WinCount = context.Params.GetRequiredInt("winCount")
			player.Cards[i].LeaderWinCount = context.Params.GetRequiredInt("leaderWinCount")

			player.Save(context.DB)
		}
	}
	
	context.Redirect(fmt.Sprintf("/admin/users/edit?userId=%s", player.UserID.Hex()), 302)
}

func DeleteCard(context *system.Context) {
	// parse parameters
	playerId := context.Params.GetRequiredID("playerId")
	cardId := data.DataId(context.Params.GetRequiredInt("card"))

	player, err := models.GetPlayerById(context.DB, playerId)
	util.Must(err)

	// make sure player will maintain min cards
	if len(player.Cards) > 9 {
		// remove card from inventory
		removed := false
		for i, card := range player.Cards {
			if card.DataID == cardId {
				player.Cards = append(player.Cards[:i], player.Cards[i + 1:]...)
				removed = true
				break
			}
		}

		if removed {
			// update decks
			for i, deck := range player.Decks {
				if deck.LeaderCardID == cardId {
					replaceCard(player, i, -1)
				} else {
					for j, deckCardId := range deck.CardIDs {
						if deckCardId == cardId {
							replaceCard(player, i, j)
							break
						}
					}
				}
			}

			// update DB
			player.Save(context.DB)
		}
	} else {
		context.Fail("Must maintain minimum of 9 cards for each player")
	}

	context.Redirect(fmt.Sprintf("/admin/users/edit?userId=%s", player.UserID.Hex()), 302)
}

func replaceCard(player *models.Player, deckIndex int, cardIndex int) {
	deck := &player.Decks[deckIndex]

	// iterate through card inventory
	for _, card := range player.Cards {
		if card.DataID == deck.LeaderCardID {
			continue
		}

		// check if card is already in deck
		found := false
		for _, deckCardId := range deck.CardIDs {
			if deckCardId == card.DataID {
				found = true
				break
			}
		}

		// if not in deck, then replace
		if !found {
			if cardIndex < 0 {
				deck.LeaderCardID = card.DataID
			} else {
				deck.CardIDs[cardIndex] = card.DataID
			}
			break
		}
	}
}