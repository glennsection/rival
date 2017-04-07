package models

import (
	"time"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"golang.org/x/crypto/bcrypt"
)

const userCollectionName = "users"

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Username string        `bson:"un" json:"username"`
	Password []byte        `bson:"ps" json:"-"`
	Email    string        `bson:"em,omitempty" json:"email"`
	Inserted time.Time     `bson:"ti" json:"inserted"`
	Login    time.Time     `bson:"tl" json:"login"`
}

func (user *User) HashPassword(password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		panic(fmt.Sprintf("Couldn't hash password: %v", err))
	}

	user.Password = hash
}

func InsertUser(database *mgo.Database, username string, password string) (err error) {
	// check existing user
	var user *User
	user, err = GetUserByUsername(database, username)
	if user != nil {
		panic(fmt.Sprintf("User already exists with username: %s", username))
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to access user database: %v", err))
	}

	// create user
	user = &User {
		Username: username,
	}
	user.HashPassword(password)

	user.ID = bson.NewObjectId()
	user.Inserted = time.Now()
	user.Login = time.Now()
	return database.C(userCollectionName).Insert(user)
}

func GetUserByUsername(database *mgo.Database, username string) (user *User, err error) {
	err = database.C(userCollectionName).Find(bson.M { "un": username } ).One(&user)
	return
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
