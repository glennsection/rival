package models

import (
	"encoding/json"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const playerCollectionName = "players"

type Player struct {
	ID               bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID           bson.ObjectId `bson:"us" json:"-"`
	StandardCurrency int           `bson:"cs" json:"standardCurrency"`
	PremiumCurrency  int           `bson:"cp" json:"premiumCurrency"`
	XP               int           `bson:"xp" json:"xp"`
	Cards            []Card        `bson:"cd" json:"cards"`
	Decks            []Deck        `bson:"ds" json:"decks"`
	CurrentDeck      int           `bson:"dc" json:"currentDeck"`
	Tomes            []Tome        `bson:"tm" json:"tomes"`
}

func ensureIndexPlayer(database *mgo.Database) {
	c := database.C(playerCollectionName)

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

func UpdatePlayer(database *mgo.Database, userId bson.ObjectId, data string) (err error) {
	// find existing player data
	player, err := GetPlayerByUser(database, userId)
	if err != nil {
		log.Println(err)
	}
	
	// initialize new player if none exists
	if player == nil {
		player = &Player {}
		player.Initialize()
		player.ID = bson.NewObjectId()
		player.UserID = userId
		
		err = nil
	}
	
	// parse updated data
	err = json.Unmarshal([]byte(data), &player)
	if err != nil {
		panic(err)
	}
	
	// update database
	_, err = database.C(playerCollectionName).Upsert(bson.M { "us": player.UserID }, player)
	return
}

func GetPlayerByUser(database *mgo.Database, userId bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = database.C(playerCollectionName).Find(bson.M { "us": userId } ).One(&player)
	return
}