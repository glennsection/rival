package models

import (
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const playerCollectionName = "players"

type Player struct {
	ID               bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserID           bson.ObjectId `bson:"us" json:"userId"`
	StandardCurrency int           `bson:"cs" json:"standardCurrency"`
	PremiumCurrency  int           `bson:"cp" json:"premiumCurrency"`
	XP               int           `bson:"xp" json:"xp"`
	Cards            []Card        `bson:"cd" json:"cards"`
	Decks            []Deck        `bson:"ds" json:"decks"`
	CurrentDeck      int           `bson:"dc" json:"currentDeck"`
}

func ParsePlayer(data string) (player *Player, err error) {
	player = &Player {}

	// parse json data
	err = json.Unmarshal([]byte(data), &player)
	if err != nil {
		panic(err)
	}

	return
}

func SetPlayer(database *mgo.Database, player *Player) (err error) {
	// collection := database.C(playerCollectionName)
	// previousPlayer, _ := GetPlayerByUser(database, player.UserID)

	// if previousPlayer != nil {
	// 	collection.
	// } else {
	// 	player.ID = bson.NewObjectId()
	// 	return collection.Insert(player)
	// }
	_, err = database.C(playerCollectionName).Upsert(bson.M { "us": player.UserID }, player)
	return
}

func GetPlayerByUser(database *mgo.Database, userId bson.ObjectId) (player *Player, err error) {
	err = database.C(playerCollectionName).Find(bson.M { "us": userId } ).One(&player)
	return
}