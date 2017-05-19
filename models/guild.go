package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
)

const GuildCollectionName = "guilds"

type Guild struct {
	ID              	bson.ObjectId `bson:"_id,omitempty" json:"-"`
	OwnerID        	 	bson.ObjectId `bson:"ow" json:"-"`
	Name                string        `bson:"nm" json:"name"`
	XP                  int           `bson:"xp" json:"xp"`
	Rating          	int           `bson:"rt" json:"rating"`

	WinCount       		int           `bson:"wc" json:"winCount"`
	LossCount       	int           `bson:"lc" json:"lossCount"`
	MatchCount       	int           `bson:"mc" json:"matchCount"`
}

func ensureIndexGuild(database *mgo.Database) {
	c := database.C(GuildCollectionName)

	// owner index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "ow" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}))
}

func GetGuildById(database *mgo.Database, id bson.ObjectId) (guild *Guild, err error) {
	// find guild by ID
	err = database.C(GuildCollectionName).Find(bson.M { "_id": id } ).One(&guild)
	return
}

func GetGuildByOwner(database *mgo.Database, ownerId bson.ObjectId) (guild *Guild, err error) {
	// find guild by owner ID
	err = database.C(GuildCollectionName).Find(bson.M { "ow": ownerId } ).One(&guild)
	return
}

func (guild *Guild) initialize() {
	guild.XP = 0
	guild.Rating = 0
	guild.WinCount = 0
	guild.LossCount = 0
	guild.MatchCount = 0
}

func CreateGuild(ownerID bson.ObjectId, name string) (guild *Guild) {
	guild = &Guild {}
	guild.initialize()

	guild.OwnerID = ownerID
	guild.Name = name
	return
}

func (guild *Guild) Save(database *mgo.Database) (err error) {
	if !guild.ID.Valid() {
		guild.ID = bson.NewObjectId()
	}

	// update entire guild to database
	_, err = database.C(GuildCollectionName).Upsert(bson.M { "_id": guild.ID }, guild)
	return
}

func (guild *Guild) Delete(database *mgo.Database) (err error) {
	// delete guild from database
	return database.C(GuildCollectionName).Remove(bson.M { "_id": guild.ID })
}
