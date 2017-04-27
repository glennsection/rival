package system

import (
	"bytes"
	"time"
	"fmt"
	"sync"
	"strings"
	"net/http"
	"net/http/httputil"
	"net/url"
	"html"
	"encoding/json"
	"runtime/debug"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/sessions"

	"bloodtales/config"
	"bloodtales/models"
	"bloodtales/log"
)

type Context struct {
	Application *Application         `json:"-"`
	Config      *config.Config       `json:"-"`
	DBSession   *mgo.Session         `json:"-"`
	DB          *mgo.Database        `json:"-"`
	Session     *sessions.Session    `json:"-"`
	Cache       *Cache               `json:"-"`
	Request     *http.Request        `json:"-"`
	Params      *Stream              `json:"-"`
	User		*models.User         `json:"-"`
	Template    string               `json:"-"`

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

	// create concurrent cookie session
	cookieSession, _ := application.cookies.Get(r, "session")

	// get concurrent cache connection
	cache := application.GetCache()

	// create context
	context := &Context {
		Application: application,
		Config: &application.Config,
		DBSession: contextDBSession,
		DB: contextDB,
		Session: cookieSession,
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

func (context *Context) SaveSession() error {
	return context.Session.Save(context.Request, context.responseWriter)
}

func (context *Context) Write(p []byte) (n int, err error) {
	// remember custom was response written
	context.responseWritten = true
	return context.responseWriter.Write(p)
}

func (source ContextStreamSource) Has(name string) bool {
	// check bindings
	source.mutex.RLock()
	defer source.mutex.RUnlock()
	_, ok := source.bindings[name]
	return ok
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

func (context *Context) GetPlayerName(playerID bson.ObjectId) string {
	key := fmt.Sprintf("PlayerName:%s", playerID.Hex())
	name := "[None]"

	if context.Cache.Has(key) {
		name = context.Cache.GetString(key, "[None]")
	}

	if name == "[None]" {
		player, err := models.GetPlayerById(context.DB, playerID)
		if err == nil && player != nil {
			context.Cache.Set(key, player.Name)
			name = player.Name
		}
	}
	return name
}

func (context *Context) Message(message string) {
	context.Messages = append(context.Messages, message)

	// TODO - add to session flashes
}

func (context *Context) Messagef(message string, params ...interface{}) {
	context.Messages = append(context.Messages, fmt.Sprintf(message, params...))

	// TODO - add to session flashes
}

func (context *Context) Fail(message string) {
	context.Success = false
	context.Message(message)
}

func (context *Context) Redirect(path string, responseCode int) {
	http.Redirect(context.responseWriter, context.Request, path, responseCode)
}

func (context *Context) BeginRequest(authType AuthenticationType, template string) {
	// remember defined template
	context.Template = template

	// initial request logging
	switch context.Config.Logging.Requests {
	case config.BriefLogging:
		// log basic request info with truncated query
		query := context.Request.URL.RawQuery
		if query != "" {
			query, _ = url.QueryUnescape(context.Request.URL.RawQuery)
			query = "?" + strings.Replace(query, "\r\n", "", -1)
		}
		message := log.Sprintf("[cyan]Request received: %v/%v%v[-]", context.Request.Host, context.Request.URL.Path, query)
		if len(message) > 472 {
			message = message[:472] + "..."
		}

		log.Println(message)
	case config.FullLogging:
		// get formatted request dump to log
		dump, _ := httputil.DumpRequest(context.Request, true)

		log.Printf("[cyan]Request received: %q[-]", dump)
	}

	// authentication
	err := context.authenticate(authType)
	if err != nil {
		log.Errorf("Failed to authenticate user: %v", err)
		context.Fail("Failed to authenticate user")

		context.Redirect("/admin", 302)
	}
}

func (context *Context) EndRequest(startTime time.Time) {
	// cleanup
	defer context.DBSession.Close()

	// handle any panic errors during request
	redirected := false
	var caughtErr interface{}
	if caughtErr = recover(); caughtErr != nil {
		// update context for failure
		context.Fail(fmt.Sprintf("%v", caughtErr))

		if context.Template != "" {
			context.Redirect(fmt.Sprintf("/error?message=%v", caughtErr), 302) // TODO - can remove parameter once session flashes are working
			redirected = true
		}
	}
	// catch any panics in this function
	defer func() {
		if templateErr := recover(); templateErr != nil {
			log.Errorf("Occurred during last request: %v", templateErr)
			log.Printf("[red]%v[-]", string(debug.Stack()))

			if context.Template != "" {
				context.Redirect(fmt.Sprintf("/error?message=%v", templateErr), 302) // TODO - can remove parameter once session flashes are working
			}
		}
	}()

	if !redirected {
		// check if any custom response was written
		if context.responseWritten {
			// nothing left to do...
		} else if context.Template != "" {
			// HTML escape messages
			for i, message := range context.Messages {
				context.Messages[i] = html.EscapeString(message)
			}

			// render template to buffer
			var output bytes.Buffer
			err := context.Application.templates.ExecuteTemplate(&output, context.Template, context)

			var responseString string
			if err == nil {
				// convert template output to string
				responseString = output.String()

				// write response to stream
				fmt.Fprint(context.responseWriter, responseString)
			} else {
				// respond with error
				responseString = fmt.Sprintf("Processing template (%v): %v", context.Template, err)

				log.Error(responseString)
				context.Redirect(fmt.Sprintf("/error?message=%s", responseString), 302) // TODO - can remove parameter once session flashes are working
			}
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