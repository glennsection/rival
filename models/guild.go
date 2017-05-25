package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/data"
)

const GuildCollectionName = "guilds"

type GuildRole int
const (
	GuildMember GuildRole = iota
	GuildElder
	GuildCaptain
	GuildOwner
)

type Guild struct {
	ID              	bson.ObjectId   `bson:"_id,omitempty" json:"-"`
	OwnerID        	 	bson.ObjectId   `bson:"ow" json:"-"`
	Name                string          `bson:"nm" json:"name"`
	XP                  int             `bson:"xp" json:"xp"`
	Rating          	int             `bson:"rt" json:"rating"`
	MemberCount         int             `bson:"ms" json"-"`

	WinCount       		int             `bson:"wc" json:"winCount"`
	LossCount       	int             `bson:"lc" json:"lossCount"`
	MatchCount       	int             `bson:"mc" json:"matchCount"`
}

// client model
type GuildClientAlias Guild
type GuildClient struct {
	Members             []*PlayerClient `json:"members"`

	*GuildClientAlias
}

func (guild *Guild) CreateGuildClient(database *mgo.Database) (client *GuildClient, err error) {
	// get member players
	var memberPlayers []*Player
	err = database.C(PlayerCollectionName).Find(bson.M { "gd": guild.ID } ).All(&memberPlayers)
	if err != nil {
		return
	}

	// create client member players
	var members []*PlayerClient
	for _, memberPlayer := range memberPlayers {
		var member *PlayerClient
		member, err = memberPlayer.CreatePlayerClient(database)
		if err != nil {
			return
		}

		members = append(members, member)
	}

	// create client model
	client = &GuildClient {
		Members: members,

		GuildClientAlias: (*GuildClientAlias)(guild),
	}
	return
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

func CreateGuild(database *mgo.Database, owner *Player, name string) (guild *Guild, err error) {
	// init guild
	guild = &Guild {}
	guild.initialize()

	// set owner and name
	guild.OwnerID = owner.ID
	guild.Name = name
	guild.MemberCount = 1

	// save guild
	err = guild.Save(database)
	if err != nil {
		return
	}

	// set guild and role for player
	owner.GuildID = guild.ID
	owner.GuildRole = GuildOwner
	err = owner.Save(database)
	if err != nil {
		return
	}

	// set dirty for return data
	owner.SetDirty(PlayerDataMask_Guild)
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

func (guild *Guild) GetLevel() int {
	// TODO - different function for guilds?
	return data.GetAccountLevel(guild.XP)
}