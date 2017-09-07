package models

import (
	"time"
	"fmt"
	"strings"

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

// user device data
type Device struct {
	ID           string        `bson:"id" json:"id"`
	Credentials  []Credential  `bson:"cds,omitempty" json:"-"`
}

// user data
type User struct {
	ID              bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Admin           bool          `bson:"ad" json:"-"`
	Username        string        `bson:"un,omitempty" json:"-"`
	Password        []byte        `bson:"ps,omitempty" json:"-"`
	Email           string        `bson:"em,omitempty" json:"-"`
	CreatedTime     time.Time     `bson:"t0" json:"created"`
	TimeZone 		string 		  `bson:"tz" json:"timeZone"` 	

	Credentials     []Credential  `bson:"cds,omitempty" json:"-"`
	Devices         []Device      `bson:"dvs" json"-"`

	Tag             string        `bson:"tag" json:"tag"`
	Name            string        `bson:"nm" json"name"`

	LastSocketTime  time.Time     `bson:"lsk" json:"-"`
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

	// Devices index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "dvs.id" },
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

func InsertUserWithDevice(context *util.Context, uuid string, credentials []Credential) (user *User, err error) {
	// TODO check for existing user?  - prevented by database?

	// generate unique user tag
	tag := util.GenerateTag()

	// create user
	user = &User {
		ID: bson.NewObjectId(),
		Credentials: credentials,
		Devices: []Device {
			Device { ID: uuid },
		},
		Tag: tag,
		CreatedTime: time.Now().UTC(),
		TimeZone: context.Params.GetString("timeZone", "UTC"),
	}

	err = context.DB.C(UserCollectionName).Insert(user)
	return
}

func InsertUserWithUsername(context *util.Context, username string, password string) (user *User, err error) {
	// TODO check for existing user?  - prevented by database

	return InsertUserWithUsernameAndDatabase(context.DB, username, password, context.Params.GetString("timeZone", "UTC"), false)
}

func InsertUserWithUsernameAndDatabase(database *mgo.Database, username string, password string, timeZone string, admin bool) (user *User, err error) {
	// create user
	user = &User {
		ID: bson.NewObjectId(),
		Username: username,
		Name: username,
		Admin: admin,
		CreatedTime: time.Now().UTC(),
		TimeZone: timeZone,
	}
	user.HashPassword(password)

	err = database.C(UserCollectionName).Insert(user)
	return
}

func GetUserById(context *util.Context, id bson.ObjectId) (user *User, err error) {
	err = context.DB.C(UserCollectionName).Find(bson.M { "_id": id } ).One(&user)
	return
}

func GetUserByDevice(context *util.Context, uuid string) (user *User, err error) {
	err = context.DB.C(UserCollectionName).Find(bson.M { "dvs.id": uuid } ).One(&user)
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

func (user *User) AppendCredentials(context *util.Context, credentials []Credential) error {
	appended := false

	for _, credential := range credentials {
		found := false

		for _, previousCredential := range user.Credentials {
			if previousCredential.Provider == credential.Provider && previousCredential.ID == credential.ID {
				found = true
				break
			}
		}

		if !found {
			user.Credentials = append(user.Credentials, credential)
		}
	}

	if appended {
		return user.Save(context)
	}
	return nil
}

func (user *User) GetCredentialsString() (string) {
	var deviceStrings []string
	i := 0

	if len(user.Credentials) > 0 {
		deviceStrings = make([]string, len(user.Devices) + 1)
		deviceStrings[i] = fmt.Sprintf("[*: %s]", formatCredentialsString(user.Credentials))
		i += 1
	} else {
		deviceStrings = make([]string, len(user.Devices))
	}

	for _, device := range user.Devices {
		if len(device.Credentials) > 0 {
			deviceCredentials := formatCredentialsString(device.Credentials)
			deviceStrings[i] = fmt.Sprintf("[%s: %s]", device.ID, deviceCredentials)
		} else {
			deviceStrings[i] = fmt.Sprintf("[%s]", device.ID)
		}
		i += 1
	}
	return strings.Join(deviceStrings, ", ")
}

func formatCredentialsString(credentials []Credential) (string) {
	credentialStrings := make([]string, len(credentials))
	for i, credential := range credentials {
		credentialStrings[i] = fmt.Sprintf("%s: %s", credential.Provider, credential.ID)
	}
	return strings.Join(credentialStrings, ", ")
}
