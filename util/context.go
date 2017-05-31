package util

import (
	"bytes"
	"time"
	"fmt"
	"strings"
	"net/http"
	"net/http/httputil"
	"net/url"
	"html"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/config"
	"bloodtales/log"
)

type Context struct {
	UserID          bson.ObjectId          `json:"-"`
	DB              *mgo.Database          `json:"-"`
	Session         *Session               `json:"-"`
	Cache           *Cache                 `json:"-"`
	Client          *Client                `json:"-"`
	Request         *http.Request          `json:"-"`
	ResponseWriter  http.ResponseWriter    `json:"-"`
	Params          *Stream                `json:"-"`
	Template        string                 `json:"-"`

	Token           string                 `json:"token"`
	Success         bool                   `json:"success"`
	Messages        []string               `json:"messages"`
	Data            map[string]interface{} `json:"data"`

	// internal
	responseWritten bool
}

func CreateContext(w http.ResponseWriter, r *http.Request) *Context {
	// cookies session
	session := GetSession(w, r)

	// create context
	context := &Context {
		DB: GetDatabaseConnection(),
		Session: session,
		Cache: GetCacheConnection(),
		Client: LoadClient(session),
		Request: r,
		ResponseWriter: w,
		Params: NewParamsStream(r),

		Token: "",
		Success: true,
		Data: map[string]interface{} {},
	}

	return context
}

func (context *Context) Write(p []byte) (n int, err error) {
	// remember custom was response written
	context.responseWritten = true
	return context.ResponseWriter.Write(p)
}

func (context *Context) SetResponseWritten() {
	context.responseWritten = true
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

func (context *Context) SetData(name string, value interface{}) {
	context.Data[name] = value
}

func (context *Context) Redirect(path string, responseCode int) {
	context.responseWritten = true
	http.Redirect(context.ResponseWriter, context.Request, path, responseCode)
}

func (context *Context) BeginRequest(template string) {
	// remember defined template
	context.Template = template

	// initial request logging
	switch config.Config.Logging.Requests {
	case config.BriefLogging:
		// log basic request info with truncated query
		query := context.Request.URL.RawQuery
		if query != "" {
			query, _ = url.QueryUnescape(context.Request.URL.RawQuery)
			query = "?" + strings.Replace(query, "\r\n", "", -1)
		}
		message := log.Sprintf("[cyan]Request received: %v%v%v[-]", context.Request.Host, context.Request.URL.Path, query)
		if len(message) > 472 {
			message = message[:472] + "..."
		}

		log.Println(message)
	case config.FullLogging:
		// get formatted request dump to log
		dump, _ := httputil.DumpRequest(context.Request, true)

		log.Printf("[cyan]Request received: %q[-]", dump)
	}
}

func (context *Context) EndRequest(startTime time.Time) {
	// cleanup connection
	defer context.DB.Session.Close()

	// handle any panics or web errors, which occurred during request
	var caughtErr interface{}
	if caughtErr = recover(); caughtErr != nil {
		// update context for failure
		context.Fail(fmt.Sprintf("%v", caughtErr))
	}
	if !context.Success && context.Template != "" {
		context.Redirect(fmt.Sprintf("/error?message=%s", context.Messages[0]), 302) // TODO - can remove parameter once session flashes are working
	}

	// catch any panics occurring in this function
	defer func() {
		if templateErr := recover(); templateErr != nil {
			PrintError("Occurred during last request", templateErr)

			if context.Template != "" {
				context.Redirect(fmt.Sprintf("/error?message=%v", templateErr), 302) // TODO - can remove parameter once session flashes are working
			}
		}
	}()

	// check if any custom response was written by the handler
	if context.responseWritten {
		// nothing left to do...
	} else if context.Template != "" {
		// escape messages for HTML template
		for i, message := range context.Messages {
			context.Messages[i] = html.EscapeString(message)
		}

		// render template to buffer
		var output bytes.Buffer
		err := GetTemplates().ExecuteTemplate(&output, context.Template, context)

		var responseString string
		if err == nil {
			// convert template output to string
			responseString = output.String()

			// write response to stream
			fmt.Fprint(context.ResponseWriter, responseString)
		} else {
			// respond with error
			responseString = fmt.Sprintf("Processing template (%v): %v", context.Template, err)

			log.Error(responseString)
			context.Redirect(fmt.Sprintf("/error?message=%s", responseString), 302) // TODO - can remove parameter once session flashes are working
		}
	} else {
		// serialize API response to json
		var responseString string
		raw, err := json.Marshal(context)
		if err == nil {
			responseString = string(raw)
		} else {
			responseString = fmt.Sprintf("Marshalling response: %v", err)

			log.Error(responseString)
		}

		// write API response to stream
		fmt.Fprint(context.ResponseWriter, responseString)
	}

	// show response profiling info
	switch config.Config.Logging.Requests {
	case config.BriefLogging, config.FullLogging:
		successMessage := "Success"
		successColor := "green!"
		if context.Success == false {
			successMessage = "Failed"
			successColor = "red!"
		}
		Profile(log.Sprintf("[cyan]Request handled: %v%v ([" + successColor + "]%s[-][cyan])[-]", context.Request.Host, context.Request.URL.Path, successMessage), startTime)

		if caughtErr == nil && context.Success == false {
			log.Errorf("Request failed with: %s", context.Messages[0])
		}
	}

	// show the error caught eariler
	if caughtErr != nil {
		PrintError("Occurred during last request", caughtErr)
	}
}