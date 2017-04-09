package models

import (
	"time"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const trackingCollectionName = "trackings"

type Tracking struct {
	ID             bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID         bson.ObjectId `bson:"us" json:"-"`
	Time           time.Time     `bson:"ti" json:"time"`
	Lifetime       time.Duration `bson:"ex" json:"expires"`
	Message        string        `bson:"ms" json:"message"`
	Data           bson.M        `bson:"dt,omitempty" json:"data"`
}

func ensureIndexTracking(database *mgo.Database) {
	c := database.C(trackingCollectionName)

	index := mgo.Index {
		Key:          []string { "UserID" },
		Unique:       false,
		DropDups:     false,
		Background:   true,
		Sparse:       true,
		//ExpiresAfter: time.Duration { ... },
	}

	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func InsertTracking(database *mgo.Database, tracking *Tracking) error {
	tracking.Time = time.Now()
	return database.C(trackingCollectionName).Insert(tracking)
}

func GetTrackings(database *mgo.Database, userId bson.ObjectId) (trackings *[]Tracking, err error) {
	err = database.C(trackingCollectionName).Find(bson.M { "us": userId } ).Sort("ti").All(&trackings)
	return
}
