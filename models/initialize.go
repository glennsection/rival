package models

import (
	"io/ioutil"
	"encoding/json"

	"bloodtales/util"
)

// initialize models and collections
func init() {
	db := util.GetDatabaseConnection()
	defer db.Session.Close()

	ensureIndexUser(db)
	ensureIndexPlayer(db)
	ensureIndexTracking(db)
	ensureIndexMatch(db)
	ensureIndexNotification(db);
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
