package system

import (
	"time"

	"github.com/gorilla/websocket"

	"bloodtales/util"
	"bloodtales/log"
)

// internal constants
const (
	// time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// maximum message size allowed from peer
	maxMessageSize = 512

	// read/write buffer size
	bufferSize = 512
)

// socket message
type SocketMessage struct {
	Content       []byte
}

// socket client
type SocketClient struct {
	connection    *websocket.Conn
	send          chan SocketMessage
}

// internal globals
var (
	clients       map[*SocketClient]bool = make(map[*SocketClient]bool)
	broadcast     chan SocketMessage = make(chan SocketMessage)
	register      chan *SocketClient = make(chan *SocketClient)
	unregister    chan *SocketClient = make(chan *SocketClient)
	upgrader      websocket.Upgrader = websocket.Upgrader {
		ReadBufferSize:  bufferSize,
		WriteBufferSize: bufferSize,
	}
)

// socket broadcast handler
func init() {
	// setup route
	App.HandleAPI("/socket", NoAuthentication, socketHandler)

	// start main listening routine
	go func() {
		for {
			// listen to global channels
			select {

			case client := <-register:
				// add new client
				clients[client] = true

			case client := <-unregister:
				// remove client
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
				}

			case message := <-broadcast:
				// broadcast message to all clients
				for client := range clients {
					select {

					case client.send <- message:

					default:
						// remote client
						close(client.send)
						delete(clients, client)

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
		connection: connection,
		send:       make(chan SocketMessage),
	}

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
			// settings
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// client send channel has been closed
				err := client.connection.WriteMessage(websocket.CloseMessage, []byte {})
				if err != nil {
					log.Errorf("Socket failed to write: %v", err)
				}
				return
			}

			// write message to client connection
			err := client.connection.WriteMessage(websocket.TextMessage, message.Content)
			if err != nil {
				log.Errorf("Socket failed to write: %v", err)
			}

			// TODO - could immediately send any other queued messages in "<-client.send"
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
		// var message SocketMessage // TODO - could use JSON here...
		// err := client.connection.ReadJSON(&message)
		_, message, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Errorf("Socket failed to read: %v", err)
			}
			break
		}

		// broadcast message (TODO - we don't need to broadcast all messages; only chat, etc.)
		broadcast <- SocketMessage { Content: message }

		// HACK
		// client.send <- SocketMessage { Content: []byte("HELLO WORLD") }
	}
}

// socket client remove
func (client *SocketClient) close() {
	unregister <- client
	client.connection.Close()
}

