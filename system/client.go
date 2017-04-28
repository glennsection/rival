package system

type Client struct {
	Version       string      `json:"v"`

	// internal
	context       *Context
}

func (context *Context) loadClient() (client *Client) {
	client, ok := context.Session.Get("_client").(*Client)
	if ok == false {
		client = &Client {}
	}

	client.context = context
	return
}

func (client *Client) Save() {
	client.context.Session.Set("_client", client)
}
