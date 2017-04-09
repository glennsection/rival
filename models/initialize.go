package models

import (
	"gopkg.in/mgo.v2"
)

// initialize models and collections
func Initialize(database *mgo.Database) {
	ensureIndexUser(database)
	ensureIndexPlayer(database)
	ensureIndexTracking(database)
}

// create new player data
func (player *Player) Initialize() {
	player.StandardCurrency = 1000
	player.PremiumCurrency = 10
	player.XP = 0
	player.Cards = nil
	player.Decks = nil
	player.CurrentDeck = 0
}
