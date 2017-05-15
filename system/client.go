package system

import (
	"encoding/gob"

	"bloodtales/util"
)

type Client struct {
	Version       string      `json:"v"`

	// internal
	context       *Context
}

func (application *Application) initializeClient() {
	gob.Register(&Client {})
}

func (context *Context) loadClient() (client *Client) {
	client, ok := context.Session.Get("_client").(*Client)
	if ok == false {
		//log.Println("Creating fresh client for session")
		client = &Client {}
	}

	client.context = context
	return
}

func (client *Client) Save() {
	client.context.Session.Set("_client", client)
	util.Must(client.context.Session.Save())
}
