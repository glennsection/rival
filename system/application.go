package system

import (
	"os"
	"fmt"
	"net/http"
	"log"

	"gopkg.in/mgo.v2"
)

type Application struct {
	DBSession *mgo.Session
	DB        *mgo.Database
}

func GetEnv(name string, defaultValue string) string {
	// get environment variable or default value
	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

func GetRequiredEnv(name string) string {
	// get required environment variable
	value := os.Getenv(name)
	if value == "" {
		panic(name + " environment variable not set")
	}

	return value
}

func HandleErrors() {
	// handle any panic errors
	if err := recover(); err != nil {
		message := fmt.Sprintf("Error occurred during execution: %v", err)
		log.Println(message)
	}
}

func HandleRequestErrors(w http.ResponseWriter, r *http.Request) {
	// handle any panic errors during request
	if err := recover(); err != nil {
		message := fmt.Sprintf("Error occurred during request: %v", err)
		log.Println(message)
		fmt.Fprint(w, message)
	}
}

func (application *Application) Init() {
	// TODO - do I need to do this...?
	//gob.Register(bson.ObjectId(""))

	// connect database session
	uri := GetRequiredEnv("MONGODB_URI")
	var err error
	application.DBSession, err = mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	application.DBSession.SetSafe(&mgo.Safe{})

	// get default database
	dbname := GetRequiredEnv("MONGODB_DB")
	application.DB = application.DBSession.DB(dbname)
}

func (application *Application) Close() {
	defer HandleErrors()

	// cleanup database connection
	if application.DBSession != nil {
		application.DBSession.Close()
	}
}

func (application *Application) Handle(pattern string, handler func(http.ResponseWriter, *http.Request, *Application)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		defer HandleRequestErrors(w, r)

		handler(w, r, application)
	})
}

func (application *Application) Serve() {
	// start serving on port
	port := GetRequiredEnv("PORT")

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		panic(err)
	}
}
