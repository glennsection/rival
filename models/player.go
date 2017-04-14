package models

import (
	"encoding/json"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const PlayerCollectionName = "players"

type Player struct {
	ID               bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID           bson.ObjectId `bson:"us" json:"-"`
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

func UpdatePlayer(database *mgo.Database, user *User, data string) (err error) {
	// find existing player data
	player, err := GetPlayerByUser(database, user.ID)
	if err != nil {
		log.Println(err)
	}
	
	// initialize new player if none exists
	if player == nil {
		player = &Player {}
		player.Initialize()
		player.ID = bson.NewObjectId()
		player.UserID = user.ID
		player.Name = user.Username
		
		err = nil
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
