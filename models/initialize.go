package models

import (
	"io/ioutil"
	"encoding/json"

	"gopkg.in/mgo.v2"
)

// initialize models and collections
func Initialize(database *mgo.Database) {
	ensureIndexUser(database)
	ensureIndexPlayer(database)
	ensureIndexTracking(database)
	ensureIndexMatch(database)
	ensureIndexNotification(database);
}

// create new player data
func (player *Player) Initialize() {
	// template for initial player
	path := "./resources/models/player.json"

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, player)
}
