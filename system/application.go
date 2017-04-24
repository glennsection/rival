package system

import (
	"os"
	"time"
	"fmt"
	"net/http"
	"html/template"
	"runtime/debug"

	"gopkg.in/mgo.v2"
	"github.com/gorilla/sessions"

	"bloodtales/config"
	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/log"
)

type Application struct {
	Config		   config.Config
	Env			  *Stream

	// internal
	dbSession		*mgo.Session
	db			   *mgo.Database
	cookies		  *sessions.CookieStore
	templates		*template.Template
}

type EnvStreamSource struct {
}

func (source EnvStreamSource) Has(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

func (source EnvStreamSource) Set(name string, value interface{}) {
	if err := os.Setenv(name, value.(string)); err != nil {
		panic(err)
	}
}

func (source EnvStreamSource) Get(name string) interface{} {
	return os.Getenv(name)
}

func (application *Application) handleErrors() {
	// handle any panic errors
	if err := recover(); err != nil {
		log.Errorf("Occurred during execution: %v", err)
		log.Printf("[red]%v[-]", string(debug.Stack()))
	}
}

func (application *Application) handleProfiler(name string, elapsedTime time.Duration) {
	// application profiling handler
	log.Printf("%s [%v]", name, elapsedTime)
}

func (application *Application) Initialize() {
	log.Println("[cyan!]Starting server application...[-]")

	// load config
	configPath := "./config.json"
	err := config.Load(configPath, &application.Config)
	if err != nil {
		panic(fmt.Sprintf("Config file (%s) failed to load: %v", configPath, err))
	}

	// create environment variables stream
	application.Env = &Stream {
		source: EnvStreamSource {},
	}

	// init profiling
	HandleProfiling(application.handleProfiler)

	// init templates
	err = application.LoadTemplates()
	if err != nil {
		panic(err)
	}
	
	// connect database
	mongoURI := application.Env.GetRequiredString("MONGODB_URI")
	application.dbSession, err = mgo.Dial(mongoURI)
	if err != nil {
		panic(err)
	}
	application.dbSession.SetSafe(&mgo.Safe {})

	// get default database
	dbname := application.Env.GetRequiredString("MONGODB_DB")
	application.db = application.dbSession.DB(dbname)

	// connect to cache
	application.initializeCache()

	// init sessions
	cookieSecret := application.Config.Sessions.CookieSecret
	application.cookies = sessions.NewCookieStore([]byte(cookieSecret))
	//application.cookies.MaxAge(60 * 60 * 8) // 8 hour expiration
	//application.cookies.Options.Secure = true // secure for OAuth

	// init models using concurrent session (DB indexes, etc.)
	tempSession := application.dbSession.Copy()
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
	if application.dbSession != nil {
		application.dbSession.Close()
	}

	// cleanup cache
	application.closeCache()
}

func (application *Application) HandleAPI(pattern string, authType AuthenticationType, handler func(*Context)) {
	application.handle(pattern, authType, handler, "")
}

func (application *Application) HandleTemplate(pattern string, authType AuthenticationType, handler func(*Context), template string) {
	application.handle(pattern, authType, handler, template)
}

func (application *Application) handle(pattern string, authType AuthenticationType, handler func(*Context), template string) {
	// all template requests start here
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// create context
		context := CreateContext(application, w, r)

		// prepare request response
		defer context.EndRequest(time.Now())

		// init context handling
		context.BeginRequest(authType, template)

		// handle request
		handler(context)
	})
}

func (application *Application) Static(pattern string, path string) {
	// get static files directory
	fs := http.FileServer(http.Dir(path))

	// fix pattern
	if pattern[len(pattern) - 1] != '/' {
		pattern = fmt.Sprintf("%v/", pattern)
	}

	// server static files from directory
	http.Handle(pattern, http.StripPrefix(pattern, fs))
}

func (application *Application) Redirect(pattern string, url string, responseCode int) {
	// redirect these requests
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, responseCode)
	})
}

func (application *Application) Ignore(pattern string) {
	// ignore these requests
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
	})
}

func (application *Application) Serve() {
	// start serving on port
	port := application.Env.GetRequiredString("PORT")

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		panic(err)
	}
}
