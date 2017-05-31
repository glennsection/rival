package models

import (
	"time"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const NotificationCollectionName = "notifications"

// an action associated with notification
type NotificationAction struct {
	Name           string               `bson:"nm" json:"name"`
	Value          string               `bson:"val" json:"value"`
}

// notification database structure
type Notification struct {
	ID             bson.ObjectId        `bson:"_id,omitempty" json:"id"`
	SenderID       bson.ObjectId        `bson:"sid,omitempty" json:"-"`
	ReceiverID     bson.ObjectId        `bson:"rid,omitempty" json:"-"`
	Guild          bool                 `bson:"gd" json:"guild"`
	CreatedAt      time.Time            `bson:"t0" json:"created"`
	ExpiresAt      time.Time            `bson:"exp" json:"expires"`
	Type           string               `bson:"tp" json:"type"`
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
		Background:   true,
	}))

	// receiver index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "rid" },
		Background:   true,
	}))

	// expiration
	util.Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "exp" },
		Background:   true,
		ExpireAfter:  time.Second,
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

func (notification *Notification) Delete(database *mgo.Database) (err error) {
	return database.C(NotificationCollectionName).Remove(bson.M { "_id": notification.ID })
}

func GetNotificationById(database *mgo.Database, id bson.ObjectId) (notification *Notification, err error) {
	// find notification by ID
	err = database.C(NotificationCollectionName).Find(bson.M { "_id": id } ).One(&notification)
	return
}

func GetSentNotifications(database *mgo.Database, player *Player) (notifications []*Notification, err error) {
	// get all notifications sent from player
	err = database.C(NotificationCollectionName).Find(bson.M { "sid": player.ID } ).Sort("t0").All(&notifications)
	return
}

func GetReceivedNotifications(database *mgo.Database, player *Player, types []string) (notifications []*Notification, err error) {
	conditions := []bson.M {
		bson.M { "gd": false, "rid": player.ID, },
		bson.M { "gd": false, "rid": bson.M { "$exists": false }, },
	}

	// get guild ID
	if player.GuildID.Valid() {
		conditions = append(conditions, bson.M { "gd": true, "rid": player.GuildID })
	}

	// type filters
	if len(types) > 0 {
		conditions = append(conditions, bson.M { "tp": bson.M { "$in": types } })
	}

	// get all pending notifications sent to player
	err = database.C(NotificationCollectionName).Find(bson.M { "$or": conditions }).Sort("t0").All(&notifications)
	if err != nil {
		return
	}
	return
}

func ViewNotificationsByIds(database *mgo.Database, ids []bson.ObjectId) (err error) {
	// remove all viewed notifications that require no action from the user
	_, err = database.C(NotificationCollectionName).RemoveAll(bson.M {
		"_id": bson.M {
			"$in": ids,
		},
		"ac": bson.M {
			"$size": 0,
		},
	})
	return
}
