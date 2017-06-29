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
	Name  string `bson:"nm" json:"name"`
	Value string `bson:"val" json:"value"`
}

// notification database structure
type Notification struct {
	ID         bson.ObjectId        `bson:"_id,omitempty" json:"id"`
	SenderID   bson.ObjectId        `bson:"sid,omitempty" json:"-"`
	ReceiverID bson.ObjectId        `bson:"rid,omitempty" json:"-"`
	Guild      bool                 `bson:"gd" json:"guild"`
	CreatedAt  time.Time            `bson:"t0" json:"created"`
	ExpiresAt  time.Time            `bson:"exp" json:"expires"`
	Type       string               `bson:"tp" json:"type"`
	Image      string               `bson:"im" json:"image"`
	Message    string               `bson:"ms" json:"message"`
	Actions    []NotificationAction `bson:"ac" json:"actions"`
	Data       bson.M               `bson:"dt,omitempty" json:"data"`

	SenderName string `bson:"-" json"sender"`
}

func ensureIndexNotification(database *mgo.Database) {
	c := database.C(NotificationCollectionName)

	// sender index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string{"sid"},
		Background: true,
	}))

	// receiver index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string{"rid"},
		Background: true,
	}))

	// expiration
	util.Must(c.EnsureIndex(mgo.Index{
		Key:         []string{"exp"},
		Background:  true,
		ExpireAfter: time.Second,
	}))
}

func (notification *Notification) Save(context *util.Context) (err error) {
	// check if notification is new
	if !notification.ID.Valid() {
		notification.ID = bson.NewObjectId()
		notification.CreatedAt = time.Now()
	}

	// update in DB
	_, err = context.DB.C(NotificationCollectionName).Upsert(bson.M{"_id": notification.ID}, notification)
	return
}

func (notification *Notification) Delete(context *util.Context) (err error) {
	return context.DB.C(NotificationCollectionName).Remove(bson.M{"_id": notification.ID})
}

func GetNotificationById(context *util.Context, id bson.ObjectId) (notification *Notification, err error) {
	// find notification by ID
	err = context.DB.C(NotificationCollectionName).Find(bson.M{"_id": id}).One(&notification)
	return
}

func GetSentNotifications(context *util.Context, player *Player) (notifications []*Notification, err error) {
	// get all notifications sent from player
	err = context.DB.C(NotificationCollectionName).Find(bson.M{"sid": player.ID}).Sort("t0").All(&notifications)
	return
}

func GetReceivedNotifications(context *util.Context, player *Player, types []string) (notifications []*Notification, err error) {
	// Get player specific conditions
	conditions := []bson.M{
		bson.M{"gd": false, "rid": player.ID},
		bson.M{"gd": false, "rid": bson.M{"$exists": false}},
	}
	// If guild ID exists, add to conditions
	if player.GuildID.Valid() {
		conditions = append(conditions, bson.M{"gd": true, "rid": player.GuildID})
	}
	// Or all player conditions together
	playerConditions := bson.M{"$or": conditions}

	//Get type conditions
	var typeConditions bson.M
	if len(types) > 0 {
		typeConditions = bson.M{"tp": bson.M{"$in": types}}
	}

	// Array of conditions to be anded
	finalConditions := []bson.M{playerConditions, typeConditions}

	// get all pending notifications sent to player
	err = context.DB.C(NotificationCollectionName).Find(bson.M{"$and": finalConditions}).Sort("t0").All(&notifications)

	if err != nil {
		return
	}

	return
}

func ViewNotificationsByIds(context *util.Context, ids []bson.ObjectId) (err error) {
	// remove all viewed notifications that require no action from the user
	_, err = context.DB.C(NotificationCollectionName).RemoveAll(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
		"ac": bson.M{
			"$size": 0,
		},
	})
	return
}
