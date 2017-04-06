package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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

func InsertPlayer(database *mgo.Database, player *Player) error {
	player.ID = bson.NewObjectId()
	return database.C("players").Insert(player)
}

func GetPlayerByUser(database *mgo.Database, userId bson.ObjectId) (player *Player, err error) {
	err = database.C("players").Find(bson.M { "us": userId } ).One(&player)
	return;
}