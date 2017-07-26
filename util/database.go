package util

import (
	"time"
	"strings"
	"io/ioutil"
	"database/sql"

	"gopkg.in/mgo.v2"
	_ "github.com/lib/pq"
)

var (
	// internal
	mgoDBSession	*mgo.Session
	mgoDB			*mgo.Database

	pqDB			*sql.DB
)

func init() {
	initDatabase();
	initSQL();
}

func initDatabase() {
	mongoURI := Env.GetRequiredString("MONGODB_URI")
	dialInfo, err := mgo.ParseURL(mongoURI)
	Must(err)

	dialInfo.Timeout = 10 * time.Second
	mgoDBSession, err = mgo.DialWithInfo(dialInfo)
	Must(err)

	// set desired session properties
	mgoDBSession.SetSyncTimeout(1 * time.Minute)
	mgoDBSession.SetSocketTimeout(1 * time.Minute)
	mgoDBSession.SetMode(mgo.Monotonic, true)
	//mgoDBSession.SetSafe(&mgo.Safe {})

	// get default database
	dbname := dialInfo.Database
	mgoDB = mgoDBSession.DB(dbname)
}

func initSQL() {
	sqlURL := Env.GetString("DATABASE_URL", "")

	if sqlURL != "" {
		var err error
		pqDB, err = sql.Open("postgres", sqlURL)
		Must(err)

		// limit connections (should be based on available plan)
		pqDB.SetMaxOpenConns(Env.GetInt("SQL_MAX_CONNECTIONS", 0))
	}
}

func GetDatabaseConnection() (*mgo.Database) {
	session := mgoDBSession.Copy()
	return mgoDB.With(session)
}

func HasSQLDatabase() (bool) {
	//return false
	return pqDB != nil
}

func ExecuteSQL(path string) {
	bytes, err := ioutil.ReadFile(path)
	Must(err)

	commands := strings.Split(string(bytes), ";")

	for _, command := range commands {
		_, err = pqDB.Exec(command)
		Must(err)
	}
}

func GetSQLDatabaseConnection() (*sql.DB) {
	return pqDB
}

func CloseDatabase() {
	// close MongoDB
	if mgoDBSession != nil {
		mgoDBSession.Close()
	}

	// close SQL
	if pqDB != nil {
		pqDB.Close()
	}
}