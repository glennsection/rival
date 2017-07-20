package models

import (
	"bloodtales/util"
)

// initialize models and collections
func init() {
	// no-sql database
	db := util.GetDatabaseConnection()
	defer db.Session.Close()
	defer func() {
		// handle any panic errors
		if err := recover(); err != nil {
			util.LogError("Occurred during database initialization", err)
		}
	}()

	ensureIndexUser(db)
	ensureIndexPlayer(db)
	ensureIndexTracking(db)
	ensureIndexMatch(db)
	ensureIndexNotification(db);
	ensureIndexFriends(db);

	util.EnsureIndexFault(db);
}
