package system

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"runtime/debug"

	"bloodtales/models"
)

type Session struct {
	Application *Application         `json:"-"`
	Response    http.ResponseWriter  `json:"-"`
	Request     *http.Request        `json:"-"`

	User		*models.User         `json:"-"`
	Token       string               `json:"token"`
	Success     bool                 `json:"success"`
	Messages    []string             `json:"messages"`
	Data        interface{}          `json:"data"`
}

func CreateSession(application *Application, w http.ResponseWriter, r *http.Request) *Session {
	return &Session {
		Application: application,
		Response: w,
		Request: r,

		User: nil,
		Token: "",
		Success: true,
	}
}

func (session *Session) GetParameter(name string) string {
	return session.Request.FormValue(name)
}

func (session *Session) GetRequiredParameter(name string) string {
	value := session.GetParameter(name)
	if value == "" {
		panic(fmt.Sprintf("Request doesn't contain required parameter: %v", name))
	}

	return value
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

func (session *Session) Respond() {
	// handle any panic errors during request
	if err := recover(); err != nil {
		log.Printf("Error occurred during request: %v", err)
		debug.PrintStack()

		// update session for failure
		session.Fail(fmt.Sprintf("%v", err))
	}

	// serialize response to json
	responseString, err := json.Marshal(session)
	if err != nil {
		responseString = []byte(fmt.Sprintf("Error marshalling response: %v", err))

		log.Println(responseString)
	}

	// write response to stream
	fmt.Fprint(session.Response, string(responseString))
}