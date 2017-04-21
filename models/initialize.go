package models

import (
	"gopkg.in/mgo.v2"
)

// initialize models and collections
func Initialize(database *mgo.Database) {
	ensureIndexUser(database)
	ensureIndexPlayer(database)
	ensureIndexTracking(database)
	ensureIndexMatch(database)
}

// create new player data
func (player *Player) Initialize() {
	player.Level = 1
	player.Rank = 0
	player.Rating = 1200
	player.WinCount = 0
	player.LossCount = 0
	player.MatchCount = 0
	player.StandardCurrency = 1000
	player.PremiumCurrency = 10

	// TODO - eventually we will set all these up too...
	//player.Cards = nil
	//player.Decks = nil
	//player.CurrentDeck = 0
	//player.Tomes = nil
}
