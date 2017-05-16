package models

import (
	"time"
	"encoding/json"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const NotificationCollectionName = "notifications"

type NotificationType int
const (
	DefaultNotification NotificationType = iota
	SystemNotification
	NewsNotification
	EventNotification
	FriendNotification
	GuildNotification
)

type Notification struct {
	ID             bson.ObjectId    `bson:"_id,omitempty" json:"-"`
	SenderID       bson.ObjectId    `bson:"sid,omitempty" json:"-"`
	ReceiverID     bson.ObjectId    `bson:"rid,omitempty" json:"-"`
	CreatedAt      time.Time        `bson:"t0" json:"created"`
	ExpiresAt      time.Time        `bson:"exp" json:"expires"`
	Type           NotificationType `bson:"tp" json:"type"`
	Image          string           `bson:"im" json:"image"`
	Message        string           `bson:"ms" json:"message"`
	Data           bson.M           `bson:"dt,omitempty" json:"data"`

	SenderName     string           `bson:"-" json"sender"`
}

func ensureIndexNotification(database *mgo.Database) {
	c := database.C(NotificationCollectionName)

	// sender index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "sid" },
		Unique:       false,
		DropDups:     false,
		Background:   true,
		Sparse:       true,
	}))

	// receiver index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "rid" },
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

// custom marshalling
func (notification* Notification) MarshalJSON() ([]byte, error) {
	// get sender name
	if notification.SenderID.Valid() {
		notification.SenderName = "" // TODO
	} else {
		notification.SenderName = ""
	}
	
	// marshal to json
	return json.Marshal(notification)
}

func (notification* Notification) Save(database *mgo.Database) (err error) {
	if !notification.ID.Valid() {
		notification.ID = bson.NewObjectId()
		notification.CreatedAt = time.Now()
	}

	_, err = database.C(NotificationCollectionName).Upsert(bson.M { "_id": notification.ID }, notification)
	return
}

func GetSentNotifications(database *mgo.Database, userId bson.ObjectId) (notifications *[]Notification, err error) {
	err = database.C(NotificationCollectionName).Find(bson.M { "sid": userId } ).Sort("t0").All(&notifications)
	return
}

func GetReceivedNotifications(database *mgo.Database, userId bson.ObjectId) (notifications *[]Notification, err error) {
	err = database.C(NotificationCollectionName).Find(bson.M { "rid": userId } ).Sort("t0").All(&notifications)
	return
}
