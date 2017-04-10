package system

import (
	"os"
	"time"
	"fmt"
	"log"
	"net/http"
	"html/template"
	"runtime/debug"

	"gopkg.in/mgo.v2"
	
	"bloodtales/data"
	"bloodtales/models"
)

type Application struct {
	DBSession        *mgo.Session
	DB               *mgo.Database

	templates        *template.Template
}

func (application *Application) GetEnv(name string, defaultValue string) string {
	// get environment variable or default value
	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

func (application *Application) GetRequiredEnv(name string) string {
	// get required environment variable
	value := os.Getenv(name)
	if value == "" {
		panic(name + " environment variable not set")
	}

	return value
}

func (application *Application) handleErrors() {
	// handle any panic errors
	if err := recover(); err != nil {
		log.Printf("Error occurred during execution: %v", err)
		debug.PrintStack()
	}
}

func (application *Application) handleProfiler(name string, elapsedTime time.Duration) {
	// application profiling handler
	log.Printf("%s [%v]", name, elapsedTime)
}

func (application *Application) Initialize() {
	// init profiling
	HandleProfiling(application.handleProfiler)

	// init templates
	application.templates = template.Must(template.ParseGlob("templates/*.tmpl.html"))
	
	// connect database session
	uri := application.GetRequiredEnv("MONGODB_URI")
	var err error
	application.DBSession, err = mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	application.DBSession.SetSafe(&mgo.Safe{})

	// get default database
	dbname := application.GetRequiredEnv("MONGODB_DB")
	application.DB = application.DBSession.DB(dbname)

	// init models (FIXME - do we need to copy the session here?)
	tempSession := application.DBSession.Copy()
	defer tempSession.Close()
	models.Initialize(tempSession.DB(dbname))

	// init auth
	application.initializeAuthentication()

	// load data
	data.Load()
}

func (application *Application) Close() {
	// handle any remaining application errors
	defer application.handleErrors()

	// cleanup database connection
	if application.DBSession != nil {
		application.DBSession.Close()
	}
}

func (application *Application) Handle(pattern string, authType AuthenticationType, handler func(*Session)) {
	// all requests start here
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// prepare session
		session := CreateSession(application, w, r)

		// prepare request response
		defer session.Respond(time.Now())

		// authentication
		err := application.authenticate(session, authType)
		if err != nil {
			panic(fmt.Sprintf("Failed to authenticate user: %v", err))
		}

		// handle request
		handler(session)
	})
}

func (application *Application) Ignore(pattern string) {
	// ignore these requests
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
	})
}

func (application *Application) Serve() {
	// start serving on port
	port := application.GetRequiredEnv("PORT")

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		panic(err)
	}
}
