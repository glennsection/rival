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

func (friends* Friends) Save(database *mgo.Database) (err error) {
	// check if friends is new
	if !friends.ID.Valid() {
		friends.ID = bson.NewObjectId()
	}

	// update in DB
	_, err = database.C(FriendsCollectionName).Upsert(bson.M { "_id": friends.ID }, friends)
	return
}

func GetFriendsByPlayerId(database *mgo.Database, playerID bson.ObjectId, allowCreate bool) (friends *Friends, err error) {
	// find friends by player ID
	err = database.C(FriendsCollectionName).Find(bson.M { "pid": playerID } ).One(&friends)

	if err != nil && err.Error() == "not found" {
		err = nil

		if allowCreate {
			friends = &Friends {
				PlayerID: playerID,
			}

			err = friends.Save(database)
		}
	}

	return
}

