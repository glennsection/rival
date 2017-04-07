package models

import (
	"time"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const trackingCollectionName = "trackings"

type Tracking struct {
	UserID         bson.ObjectId `bson:"us" json:"userId"`
	Time           time.Time     `bson:"ti" json:"time"`
	Message        string        `bson:"ms" json:"message"`
	Data           bson.M        `bson:"dt,omitempty" json:"data"`
}

func InsertTracking(database *mgo.Database, tracking *Tracking) error {
	tracking.Time = time.Now()
	return database.C(trackingCollectionName).Insert(tracking)
}

func GetTrackings(database *mgo.Database, userId bson.ObjectId) (trackings *[]Tracking, err error) {
	err = database.C(trackingCollectionName).Find(bson.M { "us": userId } ).Sort("ti").All(&trackings)
	return
}
