package models

import (
	"time"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
)

type TomeState int

const (
	TomeLocked = 0
	TomeUnlocking
	TomeUnlocked
)

const tomeCollectionName = "tomes"

type Tome struct {
	ID             bson.ObjectId `bson:"_id,omitempty" json:"id"`
	DataID         data.DataId   `bson:"di" json:"dataId"`
	State          TomeState     `bson:"st" json:"state"`
	UnlockTime     time.Time     `bson:"tu" json:"unlockTime"`
}

func InsertTome(database *mgo.Database, tome *Tome) error {
	tome.ID = bson.NewObjectId()
	return database.C(tomeCollectionName).Insert(tome)
}

func GetTome(database *mgo.Database, id bson.ObjectId) (tome *Tome, err error) {
	err = database.C(tomeCollectionName).Find(bson.M { "_id": id } ).One(&tome)
	return;
}