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
	Email    string        `bson:"em" json:"email"`
	Username string        `bson:"un" json:"username"`
	Password []byte        `bson:"ps" json:"-"`
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

func InsertUser(database *mgo.Database, user *User) error {
	user.ID = bson.NewObjectId()
	user.Inserted = time.Now()
	user.Login = time.Now()
	return database.C(userCollectionName).Insert(user)
}

func GetUserByEmail(database *mgo.Database, email string) (user *User, err error) {
	err = database.C(userCollectionName).Find(bson.M { "em": email } ).One(&user)
	return;
}

func LoginUser(database *mgo.Database, email string, password string) (user *User, err error) {
	// get user
	user, err = GetUserByEmail(database, email)
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
