package admin

import (
	"fmt"
	
	// "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/data"
	"bloodtales/models"
)

func handleAdminCards(application *system.Application) {
	//handleAdminTemplate(application, "/admin/cards/edit", system.TokenAuthentication, EditCard, "card.tmpl.html")
	handleAdminTemplate(application, "/admin/cards/delete", system.TokenAuthentication, DeleteCard, "")
}

func EditCard(context *system.Context) {
	// // parse parameters
	// userId := context.Params.GetRequiredID("userId")

	// user, err := models.GetUserById(context.DB, userId)
	// if err != nil {
	// 	panic(err)
	// }

	// player, err := models.GetPlayerByUser(context.DB, userId)
	// if err != nil {
	// 	if err.Error() != "not found" {
	// 		panic(err)
	// 	}
	// }

	// // handle request method
	// switch context.Request.Method {
	// case "POST":
	// 	userUpdated := false

	// 	tag := context.Params.GetString("tag", "")
	// 	if tag != "" {
	// 		user.Tag = tag
	// 		userUpdated = true
	// 	}

	// 	name := context.Params.GetString("name", "")
	// 	if tag != "" {
	// 		user.Name = name
	// 		userUpdated = true
	// 	}

	// 	if userUpdated {
	// 		user.Update(context.DB)
	// 	}

	// 	if player != nil {
	// 		standardCurrency := context.Params.GetInt("standardCurrency", -1)
	// 		if standardCurrency >= 0 {
	// 			player.StandardCurrency = standardCurrency
	// 		}

	// 		premiumCurrency := context.Params.GetInt("premiumCurrency", -1)
	// 		if premiumCurrency >= 0 {
	// 			player.PremiumCurrency = premiumCurrency
	// 		}

	// 		level := context.Params.GetInt("level", -1)
	// 		if level >= 0 {
	// 			player.Level = level
	// 		}

	// 		rating := context.Params.GetInt("rating", -1)
	// 		if rating >= 0 {
	// 			player.Rating = rating
	// 		}

	// 		rankPoints := context.Params.GetInt("rankPoints", -1)
	// 		if rankPoints >= 0 {
	// 			player.RankPoints = rankPoints
	// 		}

	// 		winCount := context.Params.GetInt("winCount", -1)
	// 		if winCount >= 0 {
	// 			player.WinCount = winCount
	// 		}

	// 		lossCount := context.Params.GetInt("lossCount", -1)
	// 		if lossCount >= 0 {
	// 			player.LossCount = lossCount
	// 		}

	// 		matchCount := context.Params.GetInt("matchCount", -1)
	// 		if matchCount >= 0 {
	// 			player.MatchCount = matchCount
	// 		}

	// 		player.Update(context.DB)
	// 	}

	// 	context.Message("Player updated!")
	// }
	
	// // set template bindings
	// context.Data = user
	// context.Params.Set("user", user)
	// context.Params.Set("player", player)
}

func DeleteCard(context *system.Context) {
	// parse parameters
	playerId := context.Params.GetRequiredID("playerId")
	cardId := data.DataId(context.Params.GetRequiredInt("card"))

	player, err := models.GetPlayerById(context.DB, playerId)
	if err != nil {
		panic(err)
	}

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
			player.Update(context.DB)
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