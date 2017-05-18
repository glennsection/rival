package util

import (
	"time"

	"gopkg.in/mgo.v2"
)

var (
	// internal
	dbSession	   *mgo.Session
	db			   *mgo.Database
)

func init() {
	mongoURI := Env.GetRequiredString("MONGODB_URI")
	dialInfo, err := mgo.ParseURL(mongoURI)
	Must(err)

	dialInfo.Timeout = 10 * time.Second
	dbSession, err = mgo.DialWithInfo(dialInfo)
	Must(err)

	// set desired session properties
	dbSession.SetSyncTimeout(1 * time.Minute)
	dbSession.SetSocketTimeout(1 * time.Minute)
	dbSession.SetMode(mgo.Monotonic, true)
	//dbSession.SetSafe(&mgo.Safe {})

	// get default database
	dbname := dialInfo.Database
	db = dbSession.DB(dbname)
}

func GetDatabaseConnection() (*mgo.Database) {
	session := dbSession.Copy()
	return db.With(session)
}

func CloseDatabase() {
	if dbSession != nil {
		dbSession.Close()
	}
}