package analytics

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/models"
	"bloodtales/util"
)

func (context *Context) Track(message string, data bson.M, expires time.Time) {
	user := system.GetUser(context)

	// create tracking
	tracking := &models.Tracking {
		UserID:    user.ID,
		ExpiresAt: expires,
		Message:   message,
		Data:      data,
	}

	// insert tracking
	util.Must(tracking.Insert(context.DB))
}