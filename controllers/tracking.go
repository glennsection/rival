package controllers

import (
	"time"
	"encoding/json"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handleTracking() {
	handleGameAPI("/tracking", system.TokenAuthentication, PostTracking)
}

func PostTracking(context *util.Context) {
	// parse parameters
	event := context.Params.GetRequiredString("event")
	dataJson := context.Params.GetString("data", "")
	expireAfterHours := context.Params.GetInt("expire", 0)

	// process data
	var data bson.M = nil
	if dataJson != "" {
		util.Must(json.Unmarshal([]byte(dataJson), &data))
	}

	// insert tracking
	InsertTracking(context, event, data, expireAfterHours)
}

func InsertTracking(context *util.Context, event string, data bson.M, expireAfterHours int) {
	// get user
	user := system.GetUser(context)

	// create tracking
	tracking := &models.Tracking {
		UserID: user.ID,
		Event: event,
		Data: data,
	}

	// expiration
	if expireAfterHours > 0 {
		tracking.ExpireTime = time.Now().Add(time.Hour * time.Duration(expireAfterHours))
	}

	// insert tracking
	util.Must(tracking.Insert(context))
	return
}
