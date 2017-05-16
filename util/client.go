package util

import (
	"encoding/gob"
)

type Client struct {
	Version       string      `json:"v"`

	// internal
	session       *Session
}

func init() {
	gob.Register(&Client {})
}

func LoadClient(session *Session) (client *Client) {
	client, ok := session.Get("_client").(*Client)
	if ok == false {
		//log.Println("Creating fresh client for session")
		client = &Client {}
	}

	client.session = session
	return
}

func (client *Client) Save() {
	client.session.Set("_client", client)
	Must(client.session.Save())
}
