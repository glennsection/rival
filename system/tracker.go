package system

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/models"
)

func (session *Session) Track(message string, data bson.M, lifetime time.Duration) {
	// create tracking
	tracking := &models.Tracking {
		UserID:   session.User.ID,
		Lifetime: lifetime,
		Message:  message,
		Data:     data,
	}

	// insert tracking
	if err := models.InsertTracking(session.Application.DB, tracking); err != nil {
		panic(err)
	}
}