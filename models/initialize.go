package models

import (
	"bloodtales/util"
)

// initialize models and collections
func init() {
	db := util.GetDatabaseConnection()
	defer db.Session.Close()
	defer func() {
		// handle any panic errors
		if err := recover(); err != nil {
			util.PrintError("Occurred during database initialization", err)
		}
	}()

	ensureIndexUser(db)
	ensureIndexPlayer(db)
	ensureIndexTracking(db)
	ensureIndexMatch(db)
	ensureIndexNotification(db);
	ensureIndexFriends(db);
}
