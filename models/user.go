package models

import (
	"time"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"golang.org/x/crypto/bcrypt"

	"bloodtales/util"
)

const UserCollectionName = "users"

type User struct {
	ID           bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Admin        bool          `bson:"ad" json:"admin"`
	Username     string        `bson:"un" json:"username"`
	Password     []byte        `bson:"ps" json:"-"`
	Email        string        `bson:"em,omitempty" json:"email,omitempty"`
	CreatedTime  time.Time     `bson:"t0" json:"created"`

	UUID         string        `bson:"uuid" json:"uuid,omitempty"`
	Name         string        `bson:"nm" json"name"`
	Tag          string        `bson:"tag" json:"tag,omitempty"`
}

func ensureIndexUser(database *mgo.Database) {
	c := database.C(UserCollectionName)

	// Username index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "un" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}))

	// UUID index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "uuid" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}))

	// Tag index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "tag" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}))
}

func (user *User) HashPassword(password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		panic(fmt.Sprintf("Couldn't hash password: %v", err))
	}

	user.Password = hash
}

func InsertUserWithUUID(context *util.Context, uuid string, tag string) (user *User, err error) {
	// // check existing user
	// user, _ = GetUserByUUID(context, uuid)
	// if user != nil {
	// 	panic(fmt.Sprintf("User already exists with UUID: %s", uuid))
	// }

	// create user
	user = &User {
		ID: bson.NewObjectId(),
		UUID: uuid,
		Tag: tag,
		Username: uuid,
		CreatedTime: time.Now(),
	}

	err = context.DB.C(UserCollectionName).Insert(user)
	return
}

func InsertUserWithUsername(context *util.Context, username string, password string) (user *User, err error) {
	// // check existing user
	// user, _ = GetUserByUsername(context, username)
	// if user != nil {
	// 	panic(fmt.Sprintf("User already exists with username: %s", username))
	// }

	return InsertUserWithUsernameAndDatabase(context.DB, username, password, false)
}

func InsertUserWithUsernameAndDatabase(database *mgo.Database, username string, password string, admin bool) (user *User, err error) {
	// create user
	user = &User {
		ID: bson.NewObjectId(),
		Username: username,
		Admin: admin,
		CreatedTime: time.Now(),
	}
	user.HashPassword(password)

	err = database.C(UserCollectionName).Insert(user)
	return
}

func GetUserById(context *util.Context, id bson.ObjectId) (user *User, err error) {
	err = context.DB.C(UserCollectionName).Find(bson.M { "_id": id } ).One(&user)
	return
}

func GetUserByUUID(context *util.Context, uuid string) (user *User, err error) {
	err = context.DB.C(UserCollectionName).Find(bson.M { "uuid": uuid } ).One(&user)
	return
}

func GetUserByTag(context *util.Context, tag string) (user *User, err error) {
	err = context.DB.C(UserCollectionName).Find(bson.M { "tag": tag } ).One(&user)
	return
}

func GetUserByUsername(context *util.Context, username string) (user *User, err error) {
	return GetUserByUsernameAndDatabase(context.DB, username)
}

func GetUserByUsernameAndDatabase(database *mgo.Database, username string) (user *User, err error) {
	err = database.C(UserCollectionName).Find(bson.M { "un": username } ).One(&user)
	return
}

func (user *User) Save(context *util.Context) (err error) {
	// update user in database
	_, err = context.DB.C(UserCollectionName).Upsert(bson.M { "_id": user.ID }, user)
	return
}

func (user *User) Delete(context *util.Context) (err error) {
	// delete user from database
	return context.DB.C(UserCollectionName).Remove(bson.M { "_id": user.ID })
}

func LoginUser(context *util.Context, username string, password string) (user *User, err error) {
	// get user
	user, err = GetUserByUsername(context, username)
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
