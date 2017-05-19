package system

import (
	"time"
	"fmt"
	"net/http"

	"bloodtales/log"
	"bloodtales/util"
)

type Application struct {
}

var (
	App *Application = &Application {}
)

func handleErrors() {
	// handle any panic errors
	if err := recover(); err != nil {
		util.PrintError("Occurred during execution", err)
	}
}

func handleProfiler(name string, elapsedTime time.Duration) {
	// application profiling handler
	log.Printf("%s [%v]", name, elapsedTime)
}

func init() {
	log.Println("[cyan!]Starting server application...[-]")

	// init profiling
	util.HandleProfiling(handleProfiler)
}

func (application *Application) Close() {
	// handle any remaining application errors
	defer handleErrors()

	// cleanup database connection
	util.CloseDatabase()

	// cleanup cache
	util.CloseCache()
}

func (application *Application) HandleAPI(pattern string, authType AuthenticationType, handler func(*util.Context)) {
	application.handle(pattern, authType, handler, "")
}

func (application *Application) HandleTemplate(pattern string, authType AuthenticationType, handler func(*util.Context), template string) {
	application.handle(pattern, authType, handler, template)
}

func (application *Application) handle(pattern string, authType AuthenticationType, handler func(*util.Context), template string) {
	// all template requests start here
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// create context
		context := util.CreateContext(w, r)

		// prepare request response
		defer context.EndRequest(time.Now())

		// init context handling
		context.BeginRequest(template)

		// authentication
		err := authenticate(context, authType)
		if err != nil {
			log.Errorf("Failed to authenticate user: %v", err)
			context.Fail("Failed to authenticate user")

			context.Redirect("/admin", 302)
		}

		// handle request if authenticated
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
		http.NotFound(w, r)
	})
}

func (application *Application) Serve() {
	// init templates
	util.LoadTemplates()

	// start serving on port
	port := util.Env.GetRequiredString("PORT")

	log.Printf("[cyan]Server application ready for incoming requests on port: %s[-]", port)

	util.Must(http.ListenAndServe(":" + port, nil))
}
