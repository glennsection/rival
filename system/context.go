package system

import (
	"bytes"
	"time"
	"fmt"
	"log"
	"strconv"
	"sync"
	"net/http"
	"html"
	"encoding/json"
	"runtime/debug"

	"gopkg.in/mgo.v2"

	"bloodtales/models"
)

type Context struct {
	Application *Application         `json:"-"`
	DBSession   *mgo.Session         `json:"-"`
	DB          *mgo.Database        `json:"-"`
	Request     *http.Request        `json:"-"`
	User		*models.User         `json:"-"`

	Token       string               `json:"token"`
	Success     bool                 `json:"success"`
	Messages    []string             `json:"messages"`
	Data        interface{}          `json:"data"`

	// internal
	responseWriter  http.ResponseWriter
	responseWritten bool
	bindings        map[string]interface{}
	mutex           sync.RWMutex
}

func CreateContext(application *Application, w http.ResponseWriter, r *http.Request) *Context {
	// create concurrent database session
	contextDBSession := application.dbSession.Copy()
	contextDB := application.db.With(contextDBSession)

	return &Context {
		Application: application,
		DBSession: contextDBSession,
		DB: contextDB,
		Request: r,

		// internal
		responseWriter: w,
		bindings: map[string]interface{} {},

		User: nil,
		Token: "",
		Success: true,
	}
}

func (context *Context) Write(p []byte) (n int, err error) {
	// remember custom was response written
	context.responseWritten = true
	return context.responseWriter.Write(p)
}

func (context *Context) GetParameter(name string, defaultValue string) string {
	value := context.Request.FormValue(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

func (context *Context) GetBoolParameter(name string, defaultValue bool) bool {
	value := context.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseBool(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (context *Context) GetIntParameter(name string, defaultValue int) int {
	value := context.Request.FormValue(name)
	if value != "" {
		result, err := strconv.Atoi(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (context *Context) GetFloatParameter(name string, defaultValue float64) float64 {
	value := context.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (context *Context) GetRequiredParameter(name string) string {
	value := context.Request.FormValue(name)
	if value == "" {
		panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
	}

	return value
}

func (context *Context) GetRequiredBoolParameter(name string) bool {
	value := context.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseBool(value)
		if err == nil {
			return result
		} else {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	}

	panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
}

func (context *Context) GetRequiredIntParameter(name string) int {
	value := context.Request.FormValue(name)
	if value != "" {
		result, err := strconv.Atoi(value)
		if err == nil {
			return result
		} else {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	}

	panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
}

func (context *Context) GetRequiredFloatParameter(name string) float64 {
	value := context.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return result
		} else {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	}

	panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
}

func (context *Context) GetRequiredJSONParameter(name string, result interface{}) {
	value := context.Request.FormValue(name)
	if value != "" {
		raw := []byte(value)
		err := json.Unmarshal(raw, result)
		if err != nil {
			panic(fmt.Sprintf("Request contains invalid required parameter: %v: %v", name, err))
		}
	} else {
		panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
	}
}

func (context *Context) Set(key string, value interface{}) string {
	context.mutex.Lock()
	defer context.mutex.Unlock()
	context.bindings[key] = value
	return ""
}

func (context *Context) Get(key string) interface{} {
	context.mutex.RLock()
	defer context.mutex.RUnlock()
	val := context.bindings[key]
	return val
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

func (context *Context) Respond(startTime time.Time, template string) {
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
			responseString = fmt.Sprintf("ERROR processing template (%v): %v", template, err)
			log.Println(responseString)
			//responseString = html.EscapeString(responseString)
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
			responseString = fmt.Sprintf("ERROR marshalling response: %v", err)

			log.Println(responseString)
		}

		// write response to stream
		fmt.Fprint(context.responseWriter, responseString)
	}

	// show profiling info
	Profile(fmt.Sprintf("Request: %v/%v", context.Request.Host, context.Request.URL.Path), startTime)

	if caughtErr != nil {
		log.Printf("ERROR occurred during last request: %v", caughtErr)
		debug.PrintStack()
	}
}