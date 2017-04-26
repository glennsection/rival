package models

import (
	"time"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"golang.org/x/crypto/bcrypt"
)

const UserCollectionName = "users"

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Admin    bool          `bson:"ad" json:"admin"`
	Username string        `bson:"un" json:"username"`
	Password []byte        `bson:"ps" json:"-"`
	Email    string        `bson:"em,omitempty" json:"email,omitempty"`
	Inserted time.Time     `bson:"ti" json:"inserted"`
	Login    time.Time     `bson:"tl" json:"login"`

	UUID     string        `bson:"uuid" json:"uuid,omitempty"`
	Tag      string        `bson:"tag" json:"tag,omitempty"`
	// TODO - Facebook, Game Center, Google Play IDs
}

func ensureIndexUser(database *mgo.Database) {
	c := database.C(UserCollectionName)

	// Username
	index := mgo.Index {
		Key:        []string { "un" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// UUID
	index = mgo.Index {
		Key:        []string { "uuid" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// Tag
	index = mgo.Index {
		Key:        []string { "tag" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func (user *User) HashPassword(password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		panic(fmt.Sprintf("Couldn't hash password: %v", err))
	}

	user.Password = hash
}

func InsertUserWithUUID(database *mgo.Database, uuid string, tag string) (user *User, err error) {
	// // check existing user
	// user, _ = GetUserByUUID(database, uuid)
	// if user != nil {
	// 	panic(fmt.Sprintf("User already exists with UUID: %s", uuid))
	// }

	// create user
	user = &User {
		ID: bson.NewObjectId(),
		UUID: uuid,
		Tag: tag,
		Username: uuid,
		Inserted: time.Now(),
		Login: time.Now(),
	}

	err = database.C(UserCollectionName).Insert(user)
	return
}

func InsertUserWithUsername(database *mgo.Database, username string, password string, admin bool) (user *User, err error) {
	// // check existing user
	// user, _ = GetUserByUsername(database, username)
	// if user != nil {
	// 	panic(fmt.Sprintf("User already exists with username: %s", username))
	// }

	// create user
	user = &User {
		ID: bson.NewObjectId(),
		Username: username,
		Admin: admin,
		Inserted: time.Now(),
		Login: time.Now(),
	}
	user.HashPassword(password)

	err = database.C(UserCollectionName).Insert(user)
	return
}

func GetUserById(database *mgo.Database, id bson.ObjectId) (user *User, err error) {
	err = database.C(UserCollectionName).Find(bson.M { "_id": id } ).One(&user)
	return
}

func GetUserByUUID(database *mgo.Database, uuid string) (user *User, err error) {
	err = database.C(UserCollectionName).Find(bson.M { "uuid": uuid } ).One(&user)
	return
}

func GetUserByTag(database *mgo.Database, tag string) (user *User, err error) {
	err = database.C(UserCollectionName).Find(bson.M { "tag": tag } ).One(&user)
	return
}

func GetUserByUsername(database *mgo.Database, username string) (user *User, err error) {
	err = database.C(UserCollectionName).Find(bson.M { "un": username } ).One(&user)
	return
}

func (user *User) Update(database *mgo.Database) (err error) {
	// update user in database
	_, err = database.C(UserCollectionName).Upsert(bson.M { "_id": user.ID }, user)
	return
}

func (user *User) Delete(database *mgo.Database) (err error) {
	// delete user from database
	return database.C(UserCollectionName).Remove(bson.M { "_id": user.ID })
}

func LoginUser(database *mgo.Database, username string, password string) (user *User, err error) {
	// get user
	user, err = GetUserByUsername(database, username)
	if user == nil || err != nil {
		return
	}

	// authenticate
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		user = nil
	}
	return
}
