package system

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/models"
	"bloodtales/util"
)

func (context *Context) Track(message string, data bson.M, expires time.Time) {
	// create tracking
	tracking := &models.Tracking {
		UserID:    context.User.ID,
		ExpiresAt: expires,
		Message:   message,
		Data:      data,
	}

	// insert tracking
	util.Must(tracking.Insert(context.DB))
}