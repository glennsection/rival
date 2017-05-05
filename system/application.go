package system

import (
	"os"
	"time"
	"fmt"
	"net/http"
	"html/template"

	"gopkg.in/mgo.v2"

	"bloodtales/config"
	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/log"
	"bloodtales/util"
)

type Application struct {
	Config		   config.Config
	Env			   *Stream

	// internal
	dbSession	   *mgo.Session
	db			   *mgo.Database
	templates	   *template.Template
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
		util.PrintError("Occurred during execution", err)
		//log.Errorf("Occurred during execution: %v", err)
		//log.Printf("[red]%v[-]", string(debug.Stack()))
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
	application.loadTemplates()
	
	// connect database
	application.initializeDatabase()

	// connect to cache
	application.initializeCache()

	// init sessions
	application.initializeSessions()

	// init client
	application.initializeClient()

	// init models using concurrent session (DB indexes, etc.)
	tempSession := application.dbSession.Copy()
	defer tempSession.Close()
	models.Initialize(application.db.With(tempSession))

	// init auth
	application.initializeAuthentication()

	// load data
	data.Load()

	// init player tags
	application.initializeTags()
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
		if context.Success {
			handler(context)
		}
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

	log.Printf("[cyan]Server application ready for incoming requests on port: %s[-]", port)

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		panic(err)
	}
}
