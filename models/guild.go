package models

import (
	"time"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
	"bloodtales/util"
)

const GuildCollectionName = "guilds"

type GuildRole int

const (
	GuildMember GuildRole = iota
	GuildElite
	GuildCoOwner
	GuildOwner
)

type Guild struct {
	ID          bson.ObjectId `bson:"_id,omitempty" json:"-"`
	OwnerID     bson.ObjectId `bson:"ow" json:"-"`
	Name        string        `bson:"nm" json:"name"`
	Tag         string        `bson:"tg" json:"tag"`
	Icon        string        `bson:"ic" json:"icon"`
	Description string        `bson:"ds" json:"description"`
	Private 	bool          `bson:"pr" json:"private"`
	XP          int           `bson:"xp" json:"xp"`
	Rating      int           `bson:"rt" json:"rating"`
	MemberCount int           `bson:"ms" json:"memberCount"`

	WinCount   int `bson:"wc" json:"winCount"`
	LossCount  int `bson:"lc" json:"lossCount"`
	MatchCount int `bson:"mc" json:"matchCount"`
}

// client model
type GuildClientAlias Guild
type GuildClient struct {
	Members []*PlayerClient `json:"members"`

	*GuildClientAlias
}

func GetGuildRoleName(guildRole GuildRole) string {
	switch guildRole {
	default:
		return "None"
	case GuildMember:
		return "Member"
	case GuildElite:
		return "Elite"
	case GuildCoOwner:
		return "CoOwner"
	case GuildOwner:
		return "Owner"
	}
}

func PromoteGuildRole(guildRole GuildRole) GuildRole {
	switch guildRole {
	default:
		return GuildMember
	case GuildMember:
		return GuildElite
	case GuildElite:
		return GuildCoOwner
	case GuildCoOwner:
		return GuildOwner
	}
}

func DemoteGuildRole(guildRole GuildRole) GuildRole {
	switch guildRole {
	default:
		return GuildMember
	case GuildMember:
		return GuildMember
	case GuildElite:
		return GuildMember
	case GuildCoOwner:
		return GuildElite
	}
}

func (guild *Guild) CreateGuildClient(context *util.Context) (client *GuildClient, err error) {
	// get member players
	var memberPlayers []*Player
	err = context.DB.C(PlayerCollectionName).Find(bson.M{"gd": guild.ID}).All(&memberPlayers)
	if err != nil {
		return
	}

	// create client member players
	var members []*PlayerClient
	for _, memberPlayer := range memberPlayers {
		var member *PlayerClient
		member, err = memberPlayer.GetPlayerClient(context)
		if err != nil {
			return
		}

		members = append(members, member)
	}

	// create client model
	client = &GuildClient{
		Members: members,

		GuildClientAlias: (*GuildClientAlias)(guild),
	}
	return
}

func ensureIndexGuild(database *mgo.Database) {
	c := database.C(GuildCollectionName)

	// owner index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string{"ow"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}))
}

func GetGuildById(context *util.Context, id bson.ObjectId) (guild *Guild, err error) {
	// find guild by ID
	err = context.DB.C(GuildCollectionName).Find(bson.M{"_id": id}).One(&guild)
	return
}

func GetGuildByTag(context *util.Context, tag string) (guild *Guild, err error) {
	// find guild by ID
	err = context.DB.C(GuildCollectionName).Find(bson.M{"tg": tag}).One(&guild)
	return
}

func GetGuildByOwner(context *util.Context, ownerId bson.ObjectId) (guild *Guild, err error) {
	// find guild by owner ID
	err = context.DB.C(GuildCollectionName).Find(bson.M{"ow": ownerId}).One(&guild)
	return
}

func (guild *Guild) initialize() {
	guild.XP = 0
	guild.Rating = 0
	guild.WinCount = 0
	guild.LossCount = 0
	guild.MatchCount = 0
}

func CreateGuild(context *util.Context, owner *Player, name string, iconId string, description string, private bool) (guild *Guild, err error) {
	// init guild
	guild = &Guild{}
	guild.initialize()

	// set owner and name
	guild.OwnerID = owner.ID
	guild.Name = name
	guild.MemberCount = 1
	guild.Tag = util.GenerateTag()
	guild.Icon = iconId
	guild.Description = description
	guild.Private = private

	// save guild
	err = guild.Save(context)
	if err != nil {
		return
	}

	// set guild and role for player
	owner.GuildID = guild.ID
	owner.GuildRole = GuildOwner
	owner.GuildJoinTime = time.Now().UTC()

	// set initial guild tome unlock time for owner
	owner.GuildTomeUnlockTime = util.TimeToTicks(util.GetDateInNDays(owner.TimeZone, 1))

	err = owner.Save(context)
	if err != nil {
		return
	}

	// set dirty for return data
	owner.SetDirty(PlayerDataMask_Guild, PlayerDataMask_Tomes)
	return
}

func AddMember(context *util.Context, player *Player, guild *Guild) (err error) {
	guild.MemberCount++

	if guild.MemberCount > data.GameplayConfig.GuildMemberLimit {
		err := util.NewError(fmt.Sprintf("Guild is Full. Memeber limit: %d", data.GameplayConfig.GuildMemberLimit))
		util.Must(err)
		return err
	}

	if player.GuildID.Valid() {
		err := util.NewError("Player is already in a guild")
		util.Must(err)
		return err
	}

	// set initial guild tome unlock time for owner
	player.GuildTomeUnlockTime = util.TimeToTicks(util.GetDateInNDays(player.TimeZone, 1))

	err = guild.Save(context)

	player.GuildID = guild.ID
	player.GuildRole = GuildMember
	player.GuildJoinTime = time.Now().UTC()
	player.Save(context)
	player.SetDirty(PlayerDataMask_Guild)
	return
}

func RemoveMember(context *util.Context, player *Player, guild *Guild) (err error) {
	guild.MemberCount--

	//TODO Check guild role before removing
	err = guild.Save(context)

	player.GuildID = bson.ObjectId("")

	// if Owner leaves find another player to make owner. Starting at highest rating
	if (player.GuildRole == GuildOwner && guild.MemberCount > 0) {
		fmt.Printf("Assigning a new owner")
		var memberPlayers []*Player
		err = context.DB.C(PlayerCollectionName).Find(bson.M{"gd": guild.ID}).All(&memberPlayers)
		util.Must(err)

		// Find member of highest rank first
		highestRank := GuildMember
		var highestRankPlayer *Player
		var highestRankPlayers []*Player
		for _, memberPlayer := range memberPlayers {
			if memberPlayer.ID != player.ID {
				if (highestRankPlayer == nil) {
					highestRankPlayer = memberPlayer
					highestRank = memberPlayer.GuildRole
				} else if (memberPlayer.GuildRole > highestRank) {
					highestRankPlayer = memberPlayer
					highestRank = memberPlayer.GuildRole
					highestRankPlayers = highestRankPlayers[:0] //Clear matching member rank list since we have a higher rank
					highestRankPlayers = append(highestRankPlayers, memberPlayer)
				} else if (memberPlayer.GuildRole == highestRank) {
					highestRankPlayers = append(highestRankPlayers, memberPlayer)
				}
			}
		}

		// Pick new owner based on UTC time join if multiple users of same highest rank
		if (len(highestRankPlayers) > 1) {
			var earliestJoinPlayer *Player
			var longestSeconds float64
			for _, memberPlayer := range highestRankPlayers {
				if (earliestJoinPlayer == nil) {
					earliestJoinPlayer = memberPlayer
					longestSeconds = memberPlayer.GuildJoinTime.Sub(time.Now().UTC()).Seconds()
				} else {
					diff := memberPlayer.GuildJoinTime.Sub(time.Now().UTC())
					if (diff.Seconds() > longestSeconds) {
						earliestJoinPlayer = memberPlayer
						longestSeconds = diff.Seconds()
					}
				}
			}
			highestRankPlayer = earliestJoinPlayer
		}

		if (highestRankPlayer.GuildID.Valid()) {
			fmt.Printf("new owner %s", highestRankPlayer.Name)
			highestRankPlayer.GuildRole = GuildOwner
			highestRankPlayer.Save(context)
			highestRankPlayer.SetDirty(PlayerDataMask_Guild)
		}
	}

	player.Save(context)
	player.SetDirty(PlayerDataMask_Guild)

	if guild.MemberCount <= 0 {
		guild.Delete(context)
	}

	return
}



func PromoteGuildUser(context *util.Context, player *Player, guild *Guild) (err error) {
	newGuildRole := PromoteGuildRole(player.GuildRole)

	player.GuildRole = newGuildRole
	player.Save(context)
	player.SetDirty(PlayerDataMask_Guild)

	return
}

func DemoteGuildUser(context *util.Context, player *Player, guild *Guild) (err error) {
	newGuildRole := DemoteGuildRole(player.GuildRole)

	player.GuildRole = newGuildRole
	player.Save(context)
	player.SetDirty(PlayerDataMask_Guild)

	return
}

func UpdateGuildIcon(context *util.Context, player *Player, guild *Guild, iconId string) (err error) {
	guild.Icon = iconId

	err = guild.Save(context)

	player.SetDirty(PlayerDataMask_Guild)
	return
}

func (guild *Guild) Save(context *util.Context) (err error) {
	if !guild.ID.Valid() {
		guild.ID = bson.NewObjectId()
	}

	// update entire guild to database
	_, err = context.DB.C(GuildCollectionName).Upsert(bson.M{"_id": guild.ID}, guild)
	return
}

func (guild *Guild) Delete(context *util.Context) (err error) {
	// delete guild from database
	return context.DB.C(GuildCollectionName).Remove(bson.M{"_id": guild.ID})
}

func (guild *Guild) GetLevel() int {
	// TODO - different function for guilds?
	return data.GetAccountLevel(guild.XP)
}

func (player *Player) IsInGuild() bool {
	return player.GuildID != bson.ObjectId("")
}