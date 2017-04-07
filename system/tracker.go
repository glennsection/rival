package system

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/models"
)

type Tracker struct {
	DB     *mgo.Database
	UserID bson.ObjectId
}

var tracker *Tracker = nil

func StartTracking(db *mgo.Database) {
	if tracker == nil {
		// create singleton
		tracker = &Tracker{}
	}
	tracker.DB = db
}

func Track(message string, data bson.M) {
	if tracker != nil && tracker.DB != nil {
		// create tracking
		tracking := &models.Tracking {
			UserID:  tracker.UserID, // TODO FIXME - this isn't set
			Message: message,
			Data:    data,
		}

		// insert tracking
		if err := models.InsertTracking(tracker.DB, tracking); err != nil {
			panic(err)
		}
	}
}