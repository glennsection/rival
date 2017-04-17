package system

import (
	"bytes"
	"time"
	"fmt"
	"sync"
	"net/http"
	"net/http/httputil"
	"html"
	"encoding/json"
	"runtime/debug"

	"gopkg.in/mgo.v2"

	"bloodtales/config"
	"bloodtales/models"
	"bloodtales/log"
)

type Context struct {
	Application *Application         `json:"-"`
	Config      *config.Config       `json:"-"`
	DBSession   *mgo.Session         `json:"-"`
	DB          *mgo.Database        `json:"-"`
	Cache       *Cache               `json:"-"`
	Request     *http.Request        `json:"-"`
	Params      *Stream              `json:"-"`
	User		*models.User         `json:"-"`

	Token       string               `json:"token"`
	Success     bool                 `json:"success"`
	Messages    []string             `json:"messages"`
	Data        interface{}          `json:"data"`

	// internal
	responseWriter  http.ResponseWriter
	responseWritten bool
}

type ContextStreamSource struct {
	bindings        map[string]interface{}
	mutex           sync.RWMutex
	context         *Context
}

func CreateContext(application *Application, w http.ResponseWriter, r *http.Request) *Context {
	// create concurrent database session
	contextDBSession := application.dbSession.Copy()
	contextDB := application.db.With(contextDBSession)

	// get concurrent cache connection
	cache := application.GetCache()
	defer cache.Close()

	// create context
	context := &Context {
		Application: application,
		Config: &application.Config,
		DBSession: contextDBSession,
		DB: contextDB,
		Cache: cache,
		Request: r,

		// internal
		responseWriter: w,

		User: nil,
		Token: "",
		Success: true,
	}

	// create request params stream
	context.Params = &Stream {
		source: ContextStreamSource {
			context: context,
			bindings: map[string]interface{} {},
		},
	}

	return context
}

func (context *Context) Write(p []byte) (n int, err error) {
	// remember custom was response written
	context.responseWritten = true
	return context.responseWriter.Write(p)
}

func (source ContextStreamSource) Set(name string, value interface{}) {
	// set bindings
	source.mutex.Lock()
	defer source.mutex.Unlock()
	source.bindings[name] = value
}

func (source ContextStreamSource) Get(name string) interface{} {
	// first check bindings
	source.mutex.RLock()
	defer source.mutex.RUnlock()
	if val, ok := source.bindings[name]; ok {
		return val
	}

	// then use request params
	return source.context.Request.FormValue(name)
}

func (context *Context) GetPlayer() (player *models.Player) {
	player, _ = models.GetPlayerByUser(context.DB, context.User.ID)
	return
}

func (context *Context) Message(message string) {
	context.Messages = append(context.Messages, message)
}

func (context *Context) Messagef(message string, params ...interface{}) {
	context.Messages = append(context.Messages, fmt.Sprintf(message, params...))
}

func (context *Context) Fail(message string) {
	context.Success = false
	context.Message(message)
}

func (context *Context) BeginRequest(authType AuthenticationType) {
	switch context.Config.Logging.Requests {
	case config.BriefLogging:
		// log basic request info
		log.Printf("[cyan]Request received: %v/%v?%v[-]", context.Request.Host, context.Request.URL.Path, context.Request.URL.RawQuery)
	case config.FullLogging:
		// get formatted request dump to log
		dump, err := httputil.DumpRequest(context.Request, true)
		if err != nil {
			panic(err)
		}

		log.Printf("[cyan]Request received: %q[-]", dump)
	}

	// authentication
	err := context.authenticate(authType)
	if err != nil {
		panic(fmt.Sprintf("Failed to authenticate user: %v", err))
	}
}

func (context *Context) EndRequest(startTime time.Time, template string) {
	// cleanup
	defer context.DBSession.Close()

	// handle any panic errors during request
	var caughtErr interface{}
	if caughtErr = recover(); caughtErr != nil {
		// update context for failure
		context.Fail(fmt.Sprintf("%v", caughtErr))
	}

	// check if any custom response was written
	if context.responseWritten {
		// nothing left to do...
	} else if template != "" {
		// HTML escape messages
		for i, message := range context.Messages {
			context.Messages[i] = html.EscapeString(message)
		}

		// render template to buffer
		var output bytes.Buffer
		err := context.Application.templates.ExecuteTemplate(&output, template, context)

		var responseString string
		if err == nil {
			// convert template output to string
			responseString = output.String()
		} else {
			// respond with error
			responseString = fmt.Sprintf("Processing template (%v): %v", template, err)

			log.Error(responseString)
		}

		// write response to stream
		fmt.Fprint(context.responseWriter, responseString)
	} else {
		// serialize response to json
		var responseString string
		raw, err := json.Marshal(context)
		if err == nil {
			responseString = string(raw)
		} else {
			responseString = fmt.Sprintf("Marshalling response: %v", err)

			log.Error(responseString)
		}

		// write response to stream
		fmt.Fprint(context.responseWriter, responseString)
	}

	// show response profiling info
	switch context.Config.Logging.Requests {
	case config.BriefLogging, config.FullLogging:
		successMessage := "Success"
		successColor := "green!"
		if context.Success == false {
			successMessage = "Failed"
			successColor = "red!"
		}
		Profile(log.Sprintf("[cyan]Request handled: %v/%v ([" + successColor + "]%s[-][cyan])[-]", context.Request.Host, context.Request.URL.Path, successMessage), startTime)
	}

	// show the error caught eariler
	if caughtErr != nil {
		log.Errorf("Occurred during last request: %v", caughtErr)
		log.Printf("[red]%v[-]", string(debug.Stack()))
	}
}