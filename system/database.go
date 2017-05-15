package system

import (
	"time"

	"gopkg.in/mgo.v2"

	"bloodtales/util"
)

func (application *Application) initializeDatabase() {
	mongoURI := application.Env.GetRequiredString("MONGODB_URI")
	dialInfo, err := mgo.ParseURL(mongoURI)
	util.Must(err)

	dialInfo.Timeout = 10 * time.Second
	application.dbSession, err = mgo.DialWithInfo(dialInfo)
	util.Must(err)

	// set desired session properties
	application.dbSession.SetSyncTimeout(1 * time.Minute)
	application.dbSession.SetSocketTimeout(1 * time.Minute)
	application.dbSession.SetMode(mgo.Monotonic, true)
	//application.dbSession.SetSafe(&mgo.Safe {})

	// get default database
	dbname := dialInfo.Database
	application.db = application.dbSession.DB(dbname)
}
