package system

import (
	"time"
	"fmt"
	"log"
	"strconv"
	"net/http"
	"encoding/json"
	"runtime/debug"

	"bloodtales/models"
)

type Session struct {
	Application *Application         `json:"-"`
	Request     *http.Request        `json:"-"`
	User		*models.User         `json:"-"`

	Token       string               `json:"token"`
	Success     bool                 `json:"success"`
	Messages    []string             `json:"messages"`
	Data        interface{}          `json:"data"`

	// internal
	responseWriter  http.ResponseWriter
	responseWritten bool
}

func CreateSession(application *Application, w http.ResponseWriter, r *http.Request) *Session {
	return &Session {
		Application: application,
		Request: r,

		// internal
		responseWriter: w,

		User: nil,
		Token: "",
		Success: true,
	}
}

func (session *Session) Write(p []byte) (n int, err error) {
	// remember custom was response written
	session.responseWritten = true
	return session.responseWriter.Write(p)
}

func (session *Session) GetParameter(name string, defaultValue string) string {
	value := session.Request.FormValue(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

func (session *Session) GetBoolParameter(name string, defaultValue bool) bool {
	value := session.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseBool(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (session *Session) GetIntParameter(name string, defaultValue int) int {
	value := session.Request.FormValue(name)
	if value != "" {
		result, err := strconv.Atoi(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (session *Session) GetFloatParameter(name string, defaultValue float64) float64 {
	value := session.Request.FormValue(name)
	if value != "" {
		result, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (session *Session) GetRequiredParameter(name string) string {
	value := session.Request.FormValue(name)
	if value == "" {
		panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
	}

	return value
}

func (session *Session) GetRequiredBoolParameter(name string) bool {
	value := session.Request.FormValue(name)
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

func (session *Session) GetRequiredIntParameter(name string) int {
	value := session.Request.FormValue(name)
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

func (session *Session) GetRequiredFloatParameter(name string) float64 {
	value := session.Request.FormValue(name)
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

func (session *Session) GetPlayer() (player *models.Player) {
	player, _ = models.GetPlayerByUser(session.Application.DB, session.User.ID)
	return
}

func (session *Session) Message(message string) {
	session.Messages = append(session.Messages, message)
}

func (session *Session) Messagef(message string, params ...interface{}) {
	session.Messages = append(session.Messages, fmt.Sprintf(message, params...))
}

func (session *Session) Fail(message string) {
	session.Success = false
	session.Message(message)
}

func (session *Session) Respond(startTime time.Time, template string) {
	// handle any panic errors during request
	var caughtErr interface{}
	if caughtErr = recover(); caughtErr != nil {
		// update session for failure
		session.Fail(fmt.Sprintf("%v", caughtErr))
	}

	// check if any custom response was written
	if session.responseWritten {
		// nothing left to do...
	} else if template != "" {
		// default data
		if session.Data == nil {
			session.Data = "" // TODO - find better value?
		}

		// TODO - should show caughtErr in the resulting HTML somewhere...

		// render template
		err := session.Application.templates.ExecuteTemplate(session, template, session.Data)
		if err != nil {
			responseString := fmt.Sprintf("ERROR processing template (%v): %v", template, err)

			log.Println(responseString)

			// write error response to stream
			fmt.Fprint(session.responseWriter, responseString)
		}
	} else {
		// serialize response to json
		var responseString string
		raw, err := json.Marshal(session)
		if err == nil {
			responseString = string(raw)
		} else {
			responseString = fmt.Sprintf("ERROR marshalling response: %v", err)

			log.Println(responseString)
		}

		// write response to stream
		fmt.Fprint(session.responseWriter, responseString)
	}

	// show profiling info
	Profile(fmt.Sprintf("Request: %v/%v", session.Request.Host, session.Request.URL.Path), startTime)

	if caughtErr != nil {
		log.Printf("ERROR occurred during last request: %v", caughtErr)
		debug.PrintStack()
	}
}