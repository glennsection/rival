package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
)

const cardCollectionName = "cards"

type Card struct {
	ID             bson.ObjectId `bson:"_id,omitempty" json:"id"`
	DataID         data.DataId   `bson:"di" json:"dataId"`
	Level          int           `bson:"lv" json:"level"`
	CardCount      int           `bson:"nm" json:"cardCount"`
	WinCount       int           `bson:"wc" json:"winCount"`
	LeaderWinCount int           `bson:"wl" json:"leaderWinCount"`
}

func InsertCard(database *mgo.Database, card *Card) error {
	card.ID = bson.NewObjectId()
	return database.C(cardCollectionName).Insert(card)
}

func GetCard(database *mgo.Database, id bson.ObjectId) (card *Card, err error) {
	err = database.C(cardCollectionName).Find(bson.M { "_id": id } ).One(&card)
	return;
}