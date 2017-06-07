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

// auth credentials
type Credential struct {
	Provider     string        `bson:"pv" json:"provider"`
	ID           string        `bson:"id" json:"id"`
}

type User struct {
	ID           bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Admin        bool          `bson:"ad" json:"-"`
	Username     string        `bson:"un,omitempty" json:"-"`
	Password     []byte        `bson:"ps,omitempty" json:"-"`
	Email        string        `bson:"em,omitempty" json:"-"`
	CreatedTime  time.Time     `bson:"t0" json:"created"`

	Credentials  []Credential  `bson:"cds" json:"-"`
	Tag          string        `bson:"tag" json:"tag"`
	Name         string        `bson:"nm" json"name"`
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

	// Credentials index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "cds.pv", "cds.id" },
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

func InsertUserWithCredentials(context *util.Context, credentials []Credential) (user *User, err error) {
	// TODO check for existing user?  - prevented by database

	// generate unique user tag
	tag := util.GenerateTag()

	// create user
	user = &User {
		ID: bson.NewObjectId(),
		Credentials: credentials,
		Tag: tag,
		CreatedTime: time.Now(),
	}

	err = context.DB.C(UserCollectionName).Insert(user)
	return
}

func InsertUserWithUsername(context *util.Context, username string, password string) (user *User, err error) {
	// TODO check for existing user?  - prevented by database

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

func GetUserByCredentials(context *util.Context, credentials []Credential) (user *User, err error) {
	// convert credentials into bson.M
	var query []bson.M = make([]bson.M, len(credentials), len(credentials))
	for i, credential := range credentials {
		query[i] = bson.M {
			"cds.pv": credential.Provider,
			"cds.id": credential.ID,
		}
	}

	// get user by credentials (TODO - prioritize social over device?)
	err = context.DB.C(UserCollectionName).Find(bson.M { "$or": query }).One(&user)
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
