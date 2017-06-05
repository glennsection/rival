package system

import (
	"time"

	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/websocket"

	"bloodtales/util"
	"bloodtales/log"
)

// internal constants
const (
	// debug sockets
	debugSockets = true

	// time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// maximum message size allowed from peer
	maxMessageSize = 512

	// read/write buffer size
	bufferSize = 512
)

// socket message
type SocketMessage struct {
	Message         string                  `json:"m"`
	Data            map[string]interface{}  `json:"d"`
}

// socket client
type SocketClient struct {
	userID          bson.ObjectId
	open            bool
	registered      bool
	connection      *websocket.Conn
	send            chan SocketMessage
}

// internal globals
var (
	clients         map[bson.ObjectId]*SocketClient = make(map[bson.ObjectId]*SocketClient)
	broadcast       chan SocketMessage = make(chan SocketMessage)
	register        chan *SocketClient = make(chan *SocketClient)
	unregister      chan *SocketClient = make(chan *SocketClient)
	upgrader        websocket.Upgrader = websocket.Upgrader {
		ReadBufferSize:  bufferSize,
		WriteBufferSize: bufferSize,
	}
)

// context socket send
func SocketSend(userID bson.ObjectId, message string, data map[string]interface{}) {
	if client, ok := clients[userID]; ok {
		client.send <- SocketMessage {
			Message: message,
			Data: data,
		}
	} else {
		log.Errorf("Failed to find socket connection for User ID: %v", userID)
	}
}

// socket broadcast handler
func init() {
	// handle route
	App.HandleAPI("/socket", TokenAuthentication, socketHandler)

	// start main listening routine
	go func() {
		for {
			// listen to global channels
			select {

			case client := <-register:
				// add new client
				clients[client.userID] = client

				logSocket("Registered socket for User ID: %v (%v total connections)", client.userID.Hex(), len(clients))

			case client := <-unregister:
				// unregister client
				if _, ok := clients[client.userID]; ok {
					client.unregisterClient()

					logSocket("Unregistered socket for User ID: %v", client.userID.Hex())
				}

			case message := <-broadcast:
				// broadcast message to all clients
				for userID := range clients {
					// get open client
					client := clients[userID]
					if !client.open {
						continue
					}

					// attempt to send message to client
					select {

					case client.send <- message:

					default:
						log.Errorf("Socket broadcast error for User ID: %v", client.userID)

						// unregister client
						client.unregisterClient()

					}
				}

			}
		}
	}()
}

// socket route handler
func socketHandler(context *util.Context) {
	// upgrade to web socket connection
	connection, err := upgrader.Upgrade(context.ResponseWriter, context.Request, nil)
	if err != nil {
		//http.NotFound(context.ResponseWriter, context.Request) // TODO - use API responder
		panic(err)
	}

	// create new client
	client := &SocketClient {
		userID:     context.UserID,
		open:       true,
		registered: false,
		connection: connection,
		send:       make(chan SocketMessage),
	}

	// handle connection close
	connection.SetCloseHandler(func(code int, text string) error {
		client.open = false
		unregister <- client
		return nil
	})

	// register to add client
	register <- client

	// start client routines
	go client.write()
	go client.read()

	// prevent further writes
	context.SetResponseWritten()
}

// socket client write
func (client *SocketClient) write() {
	// cleanup in case of exit
	defer client.close()

	for {
		// check for message to send
		select {

		case message, ok := <-client.send:
			// write timeout
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				if client.open {
					// client send channel has been closed
					err := client.connection.WriteMessage(websocket.CloseMessage, []byte {})
					if err != nil {
						log.Errorf("Socket failed to write close message: %v", err)
					}
				}
				return
			}

			// write message to client connection
			err := client.connection.WriteJSON(message)
			if err != nil {
				log.Errorf("Socket failed to write (%v): %v", message.Message, err)
			}

			logSocket("Socket sent message to User ID: %v (%v)", client.userID, message.Message)
		}
	}
}

// socket client read
func (client *SocketClient) read() {
	// cleanup in case of exit
	defer client.close()

	// settings
	client.connection.SetReadLimit(maxMessageSize)

	for {
		// check for message to read
		var message SocketMessage // use JSON here
		err := client.connection.ReadJSON(&message)
		// _, message, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) { // TODO? CloseNoStatusReceived, CloseAbnormalClosure
				log.Errorf("Socket failed to read: %v", err)
			}
			break
		}

		// TODO - set up a listener system to catch messages and send them along appropriately (chat, etc.)
		logSocket("Socket received message from User ID: %v (%v)", client.userID, message.Message)

		// HACK - broadcast example
		//broadcast <- SocketMessage { Content: message }
	}
}

// close socket client
func (client *SocketClient) close() {
	client.connection.Close()
}

// unregister socket client
func (client *SocketClient) unregisterClient() {
	client.registered = false

	delete(clients, client.userID)
	close(client.send)
}

// socket logging
func logSocket(message string, args ...interface{}) {
	if debugSockets {
		log.Printf("[magenta!]" + message + "[-]", args...)
	}
}