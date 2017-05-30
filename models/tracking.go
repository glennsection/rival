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
	UserID         bson.ObjectId `bson:"us,omitempty" json:"-"`
	CreatedTime    time.Time     `bson:"t0" json:"created"`
	ExpireTime     time.Time     `bson:"exp,omitempty" json:"expires"`
	Event          string        `bson:"ev" json:"event"`
	Data           bson.M        `bson:"dt,omitempty" json:"data"`
}

func ensureIndexTracking(database *mgo.Database) {
	c := database.C(TrackingCollectionName)

	// user index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "us" },
		Background:   true,
		Sparse:       true,
	}))

	// expiration
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "exp" },
		Background:   true,
		Sparse:       true,
		ExpireAfter:  1,
	}))
}

func GetTrackingById(database *mgo.Database, id bson.ObjectId) (tracking *Tracking, err error) {
	err = database.C(TrackingCollectionName).Find(bson.M { "_id": id } ).One(&tracking)
	return
}

func (tracking *Tracking) Insert(database *mgo.Database) (err error) {
	tracking.ID = bson.NewObjectId()
	tracking.CreatedTime = time.Now()
	err = database.C(TrackingCollectionName).Insert(tracking)
	return
}

func (tracking *Tracking) Delete(database *mgo.Database) (err error) {
	return database.C(TrackingCollectionName).Remove(bson.M { "_id": tracking.ID })
}

func GetTrackings(database *mgo.Database, userId bson.ObjectId) (trackings *[]Tracking, err error) {
	err = database.C(TrackingCollectionName).Find(bson.M { "us": userId } ).Sort("t0").All(&trackings)
	return
}
