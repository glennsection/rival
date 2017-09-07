package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const FriendsCollectionName = "friends"

type Friends struct {
	ID                  bson.ObjectId   `bson:"_id,omitempty"`
	PlayerID            bson.ObjectId   `bson:"pid"`
	FriendIDs           []bson.ObjectId `bson:"fd,omitempty"`
}

func ensureIndexFriends(database *mgo.Database) {
	c := database.C(FriendsCollectionName)

	// player index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "pid" },
		Unique:     true,
		DropDups:   true,
		Background: true,
	}))
}

func (friends* Friends) Save(context *util.Context) (err error) {
	// check if friends is new
	if !friends.ID.Valid() {
		friends.ID = bson.NewObjectId()
	}

	// update in DB
	_, err = context.DB.C(FriendsCollectionName).Upsert(bson.M { "_id": friends.ID }, friends)
	return
}

func GetFriendsByPlayerId(context *util.Context, playerID bson.ObjectId, allowCreate bool) (friends *Friends, err error) {
	// find friends by player ID
	err = context.DB.C(FriendsCollectionName).Find(bson.M { "pid": playerID } ).One(&friends)

	if err == mgo.ErrNotFound {
		err = nil

		if allowCreate {
			friends = &Friends {
				PlayerID: playerID,
			}

			err = friends.Save(context)
		}
	}

	return
}

