package system

import (
	//"log"
	"encoding/gob"
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
	//log.Printf("loaded client version: %s", client.Version)

	client.context = context
	return
}

func (client *Client) Save() {
	client.context.Session.Set("_client", client)
	err := client.context.Session.Save()
	if err != nil {
		panic(err)
	}
}
