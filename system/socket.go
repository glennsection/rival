package system

import (
	"time"
	"encoding/json"

	"gopkg.in/mgo.v2"
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

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// maximum message size allowed from peer
	maxMessageSize = 512

	// read/write buffer size
	bufferSize = 512
)

const SocketCollectionName = "sockets"

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

// socket model
type SocketModel struct {
	ID              bson.ObjectId			`bson:"_id,omitempty" json:"-"`
	CreatedAt       time.Time				`bson:"t0" json:"-"`
	ExpiresAt       time.Time				`bson:"exp" json:"-"`
	UserID          bson.ObjectId			`bson:"us,omitempty" json:"-"`
	Message         string					`bson:"ms" json:"message"`
	Data            map[string]interface{}  `bson:"-" json:"data"`
	JsonData        string					`bson:"dt" json:"-"`
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
func SocketSend(context *util.Context, userID bson.ObjectId, message string, data map[string]interface{}) {
	if userID.Valid() {
		if client, ok := clients[userID]; ok {
			client.send <- SocketMessage {
				Message: message,
				Data: data,
			}
		} else {
			// client may be offline
			//log.Errorf("Failed to find socket connection for User ID: %v", userID)
		}
	} else {
		for _, client := range clients {
			client.send <- SocketMessage {
				Message: message,
				Data: data,
			}
		}
	}

	// put message into DB
	rawJsonData, _ := json.Marshal(data)
	socketModel := &SocketModel {
		UserID: userID,
		Message: message,
		JsonData: string(rawJsonData),
	}
	socketModel.save(context)
}

// socket broadcast handler
func init() {
	// handle route
	App.HandleAPI("/socket/connect", TokenAuthentication, socketConnectHandler)
	App.HandleAPI("/socket/poll", TokenAuthentication, socketPollHandler)

	// indexes
	ensureIndexSocket()

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

func ensureIndexSocket() {
	// no-sql database
	db := util.GetDatabaseConnection()
	defer db.Session.Close()
	defer func() {
		// handle any panic errors
		if err := recover(); err != nil {
			util.LogError("Occurred during database initialization", err)
		}
	}()

	c := db.C(SocketCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "us", "t0" },
		Background: true,
		Sparse:     true,
	}))

	// expiration
	util.Must(c.EnsureIndex(mgo.Index{
		Key:         []string { "exp" },
		Background:  true,
		ExpireAfter: time.Second,
	}))
}

// socket route handlers
func socketConnectHandler(context *util.Context) {
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

func socketPollHandler(context *util.Context) {
	user := GetUser(context)

	// query conditions
	query := bson.M { "$and": 
		[]bson.M {
			bson.M {
				"$or": []bson.M {
					bson.M { "us": context.UserID },
					bson.M { "us": bson.M { "$exists": false} },
				},
			},
			bson.M { "t0": bson.M { "$gt": user.LastSocketTime } },
		},
	}

	// get all recently received socket messages
	var socketModels []*SocketModel
	err := context.DB.C(SocketCollectionName).Find(query).Sort("t0").All(&socketModels)
	util.Must(err)

	// check if any messages were found
	socketMessageCount := len(socketModels)
	if socketMessageCount > 0 {
		// update last socket time in user
		user.LastSocketTime = socketModels[socketMessageCount - 1].CreatedAt
		user.Save(context)
	}

	// unmarshal json data
	for _, socketModel := range socketModels {
		if socketModel.JsonData != "" {
			util.Must(json.Unmarshal([]byte(socketModel.JsonData), &socketModel.Data))
		}
	}

	// send messages to client
	context.SetData("messages", socketModels)
}

func (socketModel *SocketModel) save(context *util.Context) (err error) {
	socketModel.ID = bson.NewObjectId()
	socketModel.CreatedAt = time.Now()
	socketModel.ExpiresAt = time.Now().Add(time.Hour * time.Duration(24))

	// insert socket model in database
	err = context.DB.C(SocketCollectionName).Insert(socketModel)
	return
}

// socket client write
func (client *SocketClient) write() {
	// ping ticker
	ticker := time.NewTicker(pingPeriod)

	// cleanup in case of exit
	defer func() {
		ticker.Stop()
		client.close()
	}()

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

		case <-ticker.C:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.connection.WriteMessage(websocket.PingMessage, []byte {}); err != nil {
				return
			}

		}
	}
}

// socket client read
func (client *SocketClient) read() {
	// cleanup in case of exit
	defer client.close()

	// settings
	client.connection.SetReadLimit(maxMessageSize)
	client.connection.SetReadDeadline(time.Now().Add(pongWait))
	client.connection.SetPongHandler(func(string) error {
		client.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

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