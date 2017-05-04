package system

import (
	"time"

	"gopkg.in/mgo.v2"
)

func (application *Application) initializeDatabase() {
	mongoURI := application.Env.GetRequiredString("MONGODB_URI")
	dialInfo, err := mgo.ParseURL(mongoURI)
	if err != nil {
		panic(err)
	}
	dialInfo.Timeout = 10 * time.Second
	application.dbSession, err = mgo.DialWithInfo(dialInfo)
	// application.dbSession, err = mgo.Dial(mongoURI)
	if err != nil {
		panic(err)
	}

	// set desired session properties
	application.dbSession.SetSyncTimeout(1 * time.Minute)
	application.dbSession.SetSocketTimeout(1 * time.Minute)
	application.dbSession.SetMode(mgo.Monotonic, true)
	//application.dbSession.SetSafe(&mgo.Safe {})

	// get default database
	//dbname := application.Env.GetRequiredString("MONGODB_DB")
	dbname := dialInfo.Database
	application.db = application.dbSession.DB(dbname)
}
