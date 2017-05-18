package models

import (
	"time"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const NotificationCollectionName = "notifications"

// type of notification
type NotificationType int
const (
	DefaultNotification NotificationType = iota
	SystemNotification
	NewsNotification
	EventNotification
	FriendNotification
	GuildNotification
	ChatNotification
)

// an action associated with notification
type NotificationAction struct {
	Name           string               `bson:"nm" json:"name"`
	URL            string               `bson:"url" json:"url"`
}

// notification database structure
type Notification struct {
	ID             bson.ObjectId        `bson:"_id,omitempty" json:"-"`
	SenderID       bson.ObjectId        `bson:"sid,omitempty" json:"-"`
	ReceiverID     bson.ObjectId        `bson:"rid,omitempty" json:"-"`
	CreatedAt      time.Time            `bson:"t0" json:"created"`
	ExpiresAt      time.Time            `bson:"exp" json:"expires"`
	Type           NotificationType     `bson:"tp" json:"type"`
	Image          string               `bson:"im" json:"image"`
	Message        string               `bson:"ms" json:"message"`
	Actions        []NotificationAction `bson:"ac" json:"actions"`
	Data           bson.M               `bson:"dt,omitempty" json:"data"`

	SenderName     string               `bson:"-" json"sender"`
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

func (notification* Notification) Save(database *mgo.Database) (err error) {
	// check if notification is new
	if !notification.ID.Valid() {
		notification.ID = bson.NewObjectId()
		notification.CreatedAt = time.Now()
	}

	// update in DB
	_, err = database.C(NotificationCollectionName).Upsert(bson.M { "_id": notification.ID }, notification)
	return
}

func GetNotificationById(database *mgo.Database, id bson.ObjectId) (notifications *Notification, err error) {
	// find notification by ID
	err = database.C(NotificationCollectionName).Find(bson.M { "_id": id } ).One(&notifications)
	return
}

func GetSentNotifications(database *mgo.Database, userID bson.ObjectId) (notifications []*Notification, err error) {
	// get all notifications sent from user
	err = database.C(NotificationCollectionName).Find(bson.M { "sid": userID } ).Sort("t0").All(&notifications)
	return
}

func GetReceivedNotifications(database *mgo.Database, userID bson.ObjectId) (notifications []*Notification, err error) {
	// get all pending notifications sent to user
	err = database.C(NotificationCollectionName).Find(bson.M { "rid": userID } ).Sort("t0").All(&notifications)
	if err != nil {
		return
	}

	// remove all notifications that require no action from the user
	_, err = database.C(NotificationCollectionName).RemoveAll(bson.M {
		"rid": userID,
		"ac": bson.M {
			"$size": 0,
		},
 	})
	return
}
