package models

import (
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const PlayerCollectionName = "players"

type Player struct {
	ID               bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID           bson.ObjectId `bson:"us" json:"userId"`
	Name             string        `bson:"nm" json:"name"`
	Level            int           `bson:"lv" json:"level"`
	Rank             int           `bson:"rk" json:"rank"`
	Rating           int           `bson:"rt" json:"rating"`
	WinCount       	 int           `bson:"wc" json:"winCount"`
	LossCount        int           `bson:"lc" json:"lossCount"`
	MatchCount       int           `bson:"mc" json:"matchCount"`

	StandardCurrency int           `bson:"cs" json:"standardCurrency"`
	PremiumCurrency  int           `bson:"cp" json:"premiumCurrency"`
	Cards            []Card        `bson:"cd" json:"cards"`
	Decks            []Deck        `bson:"ds" json:"decks"`
	CurrentDeck      int           `bson:"dc" json:"currentDeck"`
	Tomes            []Tome        `bson:"tm" json:"tomes"`
}

// client model
type PlayerClientAlias Player
type PlayerClient struct {
	UserID      	string         `json:"userId"`

	*PlayerClientAlias
}

func ensureIndexPlayer(database *mgo.Database) {
	c := database.C(PlayerCollectionName)

	index := mgo.Index {
		Key:        []string { "UserID" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

// custom marshalling
func (player *Player) MarshalJSON() ([]byte, error) {
	// create client model
	client := &PlayerClient {
		UserID: player.UserID.Hex(),
		PlayerClientAlias: (*PlayerClientAlias)(player),
	}
	
	// marshal with client model
	return json.Marshal(client)
}

// custom unmarshalling
func (player *Player) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &PlayerClient {
		PlayerClientAlias: (*PlayerClientAlias)(player),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server user ID
	player.UserID = bson.ObjectId(client.UserID)

	return nil
}

func GetPlayerById(database *mgo.Database, id bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = database.C(PlayerCollectionName).Find(bson.M { "_id": id } ).One(&player)
	return
}

func GetPlayerByUser(database *mgo.Database, userId bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = database.C(PlayerCollectionName).Find(bson.M { "us": userId } ).One(&player)
	return
}

func CreatePlayer(userID bson.ObjectId, name string) (player *Player) {
	player = &Player {}
	player.Initialize()
	player.ID = bson.NewObjectId()
	player.UserID = userID
	player.Name = name
	return
}

func UpdatePlayer(database *mgo.Database, user *User, data string) (err error) {
	// find existing player data
	player, _ := GetPlayerByUser(database, user.ID)
	
	// initialize new player if none exists
	if player == nil {
		player = CreatePlayer(user.ID, user.Username)
	}
	
	// parse updated data
	err = json.Unmarshal([]byte(data), &player)
	if err != nil {
		panic(err)
	}
	
	// update database
	return player.Update(database)
}

func (player *Player) Update(database *mgo.Database) (err error) {
	// update entire player to database
	_, err = database.C(PlayerCollectionName).Upsert(bson.M { "us": player.UserID }, player)
	return
}
