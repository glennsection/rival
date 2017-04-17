package system

import (
	"os"
	"time"
	"strings"
	"fmt"
	"path/filepath"
	"net/http"
	"html/template"
	"runtime/debug"

	"gopkg.in/mgo.v2"
	
	"bloodtales/config"
	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/log"
)

type Application struct {
	Config           config.Config
	Env              *Stream

	// internal
	dbSession        *mgo.Session
	db               *mgo.Database
	templates        *template.Template
}

type EnvStreamSource struct {
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

	// connect to cache
	application.initializeCache()

	// get default database
	dbname := application.Env.GetRequiredString("MONGODB_DB")
	application.db = application.dbSession.DB(dbname)

	// init models using concurrent session (DB indexes, etc.)
	tempSession := application.dbSession.Copy()
	defer tempSession.Close()
	models.Initialize(tempSession.DB(dbname))

	// init auth
	application.initializeAuthentication()

	// load data
	data.Load()
}

func (application *Application) LoadTemplates() error {
	var templates []string

	// gather all HTML templates
	fn := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true && strings.HasSuffix(f.Name(), ".html") {
			templates = append(templates, path)
		}
		return nil
	}

	err := filepath.Walk("templates", fn)
	if err != nil {
		return err
	}

	// preload all HTML templates
	application.templates = template.Must(template.ParseFiles(templates...))
	return nil
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
		defer context.EndRequest(time.Now(), template)

		// init context handling
		context.BeginRequest(authType)

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

func (application *Application) Redirect(pattern string, url string) {
	// redirect these requests
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, 302)
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
