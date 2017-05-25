package models

import (
	"time"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const TrackingCollectionName = "trackings"

type Tracking struct {
	ID             bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID         bson.ObjectId `bson:"us" json:"-"`
	CreatedAt      time.Time     `bson:"t0" json:"created"`
	ExpiresAt      time.Time     `bson:"exp" json:"expires"`
	Message        string        `bson:"ms" json:"message"`
	Data           bson.M        `bson:"dt,omitempty" json:"data"`
}

func ensureIndexTracking(database *mgo.Database) {
	c := database.C(TrackingCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "us" },
		Unique:       false,
		DropDups:     false,
		Background:   true,
		Sparse:       true,
	}))

	// expiration
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "exp" },
		Unique:       false,
		DropDups:     false,
		Background:   true,
		Sparse:       true,
		ExpireAfter:  1,
	}))
}

func (tracking *Tracking) Insert(database *mgo.Database) (err error) {
	tracking.ID = bson.NewObjectId()
	tracking.CreatedAt = time.Now()
	err = database.C(TrackingCollectionName).Insert(tracking)
	return
}

func GetTrackings(database *mgo.Database, userId bson.ObjectId) (trackings *[]Tracking, err error) {
	err = database.C(TrackingCollectionName).Find(bson.M { "us": userId } ).Sort("ti").All(&trackings)
	return
}
