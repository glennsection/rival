package models

import (
	"time"
	"fmt"
	"math/rand"
	"encoding/json"
	"io/ioutil"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/config"
	"bloodtales/data"
	"bloodtales/util"
	"bloodtales/log"
)

const PlayerCollectionName = "players"
const MinutesToUnlockFreeTome = 5

const (
	PlayerDataMask_None util.Bits = 0x0
	PlayerDataMask_All = 0xfffffff
	PlayerDataMask_Name = 0x1
	PlayerDataMask_Currency = 0x2
	PlayerDataMask_XP = 0x4
	PlayerDataMask_Cards = 0x8
	PlayerDataMask_Deck = 0x10
	PlayerDataMask_Loadout = 0x20
	PlayerDataMask_Tomes = 0x40
	PlayerDataMask_Stars = 0x80
	PlayerDataMask_Quests = 0x100
	PlayerDataMask_Friends = 0x200
	PlayerDataMask_Guild = 0x400
)

type Player struct {
	ID                  bson.ObjectId   `bson:"_id,omitempty" json:"-"`
	UserID              bson.ObjectId   `bson:"us" json:"-"`
	LastTime            time.Time       `bson:"tz" json:"-"`
	Name                string          `bson:"-" json:"name"`
	Tag                 string          `bson:"-" json:"tag"`
	XP                  int             `bson:"xp" json:"xp"`
	RankPoints          int             `bson:"rk" json:"rankPoints"`
	Rating              int             `bson:"rt" json:"rating"`

	WinCount            int             `bson:"wc" json:"winCount"`
	LossCount           int             `bson:"lc" json:"lossCount"`
	MatchCount          int             `bson:"mc" json:"matchCount"`

	StandardCurrency    int             `bson:"cs" json:"standardCurrency"`
	PremiumCurrency     int             `bson:"cp" json:"premiumCurrency"`
	Cards               []Card          `bson:"cd" json:"cards"`
	UncollectedCards    []Card          `bson:"uc" json:"uncollectedCards"`
	Decks               []Deck          `bson:"ds" json:"decks"`
	CurrentDeck         int             `bson:"dc" json:"currentDeck"`
	Tomes               []Tome          `bson:"tm" json:"tomes"`
	ArenaPoints         int             `bson:"ap" json:"arenaPoints"`
	FreeTomes           int             `bson:"ft" json:"freeTomes"`
	FreeTomeUnlockTime  int64           `bson:"fu" json:"freeTomeUnlockTime"`

	Quests 				[]QuestSlot 	`bson:"qu" json:"quests"`

	GuildID             bson.ObjectId   `bson:"gd,omitempty" json:"-"`
	GuildRole           GuildRole       `bson:"gr,omitempty" json:"-"`

	DirtyMask           util.Bits       `bson:"-" json:"-"`

	CardsPurchased		[3]int 		    `bson:"pu" json:"-"`
	PurchaseResetTime 	int64 		    `bson:"pr" json:"-"`
}

// client model
type PlayerClient struct {
	Name                string          `json:"name"`
	Tag                 string          `json:"tag"`
	XP                  int             `json:"xp"`
	RankPoints          int             `json:"rankPoints"`
	Rating              int             `json:"rating"`

	WinCount            int             `json:"winCount"`
	LossCount           int             `json:"lossCount"`
	MatchCount          int             `json:"matchCount"`

	GuildRole           GuildRole       `json:"guildRole"`

	Online              bool            `json:"online"`
	LastOnline          int64           `json:"lastOnline"`
}

func ensureIndexPlayer(database *mgo.Database) {
	c := database.C(PlayerCollectionName)

	// username index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "us" },
		Unique:     true,
		DropDups:   true,
		Background: true,
	}))
}

func GetPlayerById(context *util.Context, id bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = context.DB.C(PlayerCollectionName).Find(bson.M { "_id": id } ).One(&player)
	return
}

func GetPlayerByUser(context *util.Context, userId bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = context.DB.C(PlayerCollectionName).Find(bson.M { "us": userId } ).One(&player)
	return
}

func (player *Player) loadDefaults() {
	// template for initial player
	path := "./resources/models/player.json"

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, player)
}

func CreatePlayer(userID bson.ObjectId) (player *Player) {
	player = &Player {}
	player.loadDefaults()

	player.Quests = make([]QuestSlot,3,3)
	for i,_ := range player.Quests {
		player.AssignRandomQuest(&(player.Quests[i]))
	}
	
	player.UserID = userID
	return
}

func (player *Player) GetPlayerClient(context *util.Context) (client *PlayerClient, err error) {
	playerUser, err := GetUserById(context, player.UserID)
	if err != nil {
		return
	}

	// check if online (TODO - better validation?)
	lastOnline := time.Now().Sub(player.LastTime)
	online := (lastOnline < time.Second * config.Config.Sessions.OfflineTimeout)

	// create player client
	client = &PlayerClient {
		Name: playerUser.Name,
		Tag: playerUser.Tag,
		XP: player.XP,
		RankPoints: player.RankPoints,
		Rating: player.Rating,

		WinCount: player.WinCount,
		LossCount: player.LossCount,
		MatchCount: player.MatchCount,

		//GuildName: ... // TODO
		GuildRole: player.GuildRole,

		Online: online,
		LastOnline: util.DurationToTicks(lastOnline),
	}
	return
}

func (player *Player) Reset(context *util.Context) (err error) {
	// reset player and update in database
	player.loadDefaults()

	return player.Save(context)
}

func ResetPlayers(context *util.Context) error {
	// TODO - this is an example of a bulk aggregate operation, but isn't fully tested...
	var result bson.D
	return context.DB.Run(bson.D {
		bson.DocElem { "update",  PlayerCollectionName },
		bson.DocElem { "updates",  []bson.M {
			bson.M {
				"q": bson.M {},
				"u": bson.M {
					"xp": 0,
					"rk": 0,
					"rt": 1200,
					"wc": 0,
					"lc": 0,
					"mc": 0,
					"ap": 0,
				},
				"multi": false,
				"upsert": false,
				"limit": 0,
			},
		} },
		bson.DocElem { "writeConcern", bson.M {
			"w": 1,
			"j": true,
			"wtimeout": 1000,
		} },
		bson.DocElem { "ordered", false },
	}, &result)
}

func (player *Player) Save(context *util.Context) (err error) {
	if !player.ID.Valid() {
		player.ID = bson.NewObjectId()
	}

	// last active time
	player.LastTime = time.Now()

	// update entire player to database
	_, err = context.DB.C(PlayerCollectionName).Upsert(bson.M { "_id": player.ID }, player)
	return
}

func (player *Player) UpdateFromJson(context *util.Context, data string) (err error) {
	// parse updated data
	err = json.Unmarshal([]byte(data), &player)
	if err == nil {
		// update database
		err = player.Save(context)
	}
	return
}

func (player *Player) Update(context *util.Context, updates bson.M) (err error) {
	// update given values
	err = context.DB.C(PlayerCollectionName).Update(bson.M { "_id": player.ID }, bson.M { "$set": updates })
	return
}

func (player *Player) Delete(context *util.Context) (err error) {
	// delete player from database
	return context.DB.C(PlayerCollectionName).Remove(bson.M { "_id": player.ID })
}

func (player *Player) GetLevel() int {
	return data.GetAccountLevel(player.XP)
}

func (player *Player) AddVictoryTome(context *util.Context) (index int, tome *Tome) {
	//first check to see if the player has an available tome slot, else return
	tome = nil
	index = -1
	for i, tomeSlot := range player.Tomes {
		if tomeSlot.State == TomeEmpty {
			index = i
			tome = &player.Tomes[i]
			break
		}
	}
	if tome == nil {
		return
	}

	//next sort our TomeData by chance
	compare := func(leftOperand *data.TomeData, rightOperand *data.TomeData) bool {
		return leftOperand.Chance > rightOperand.Chance
	}
	tomes := data.GetTomeIdsSorted(compare)

	//now roll for a tome
	rand.Seed(time.Now().UTC().UnixNano())
	roll := rand.Float64() * 100

	var accum float64
	for _, id := range tomes {
		tomeData := data.GetTome(id)
		accum += tomeData.Chance
		if roll <= accum {
			(*tome).DataID = id
			(*tome).State = TomeLocked
			(*tome).UnlockTime = util.TicksToTime(0)
			break
		}
	}
	return
}

func (player *Player) AddRewards(context *util.Context, tome *Tome) (reward *Reward, err error) {
	reward = tome.OpenTome(player.GetLevel())
	player.PremiumCurrency += reward.PremiumCurrency
	player.StandardCurrency += reward.StandardCurrency

	for i, id := range reward.Cards {
		player.AddCards(id, reward.NumRewarded[i])
	}

	err = player.Save(context)
	return
}

func (player *Player) GetDrawCount() int {
	return player.MatchCount - player.WinCount - player.LossCount
}

func (player *Player) GetWinRatio() string {
	if player.MatchCount > 0 {
		return fmt.Sprintf("%d%%", player.WinCount * 100 / player.MatchCount)
	}
	return "-"
}

func (player *Player) GetRankData() *data.RankData {
	return data.GetRank(player.RankPoints)
}

func (player *Player) GetRankTier() int {
	rank := player.GetRankData()
	if rank != nil {
		return rank.GetTier()
	}
	return 0
}

func (player *Player) GetRankName() string {
	rank := player.GetRankData()
	if rank != nil {
		tier := rank.GetTier()
		rankInTier := rank.Level - (tier - 1) * 5
		// return fmt.Sprintf("Tier %d Rank %d", tier, rankInTier)
		return fmt.Sprintf("%d-%d", tier, rankInTier)
	}
	return "Unranked"
}

func (player *Player) GetPlace(context *util.Context) int {
	return context.Cache.GetScore("Leaderboard", player.ID.Hex())
}

func (player *Player) UpdatePlace(context *util.Context) {
	matches := player.MatchCount
	if matches > 0 {
		// calculate placement score
		winsFactor := player.WinCount * 1000000 / matches
		matchesFactor := matches * 1000
		pointsFactor := player.ArenaPoints

		score := winsFactor + matchesFactor + pointsFactor
		context.Cache.SetScore("Leaderboard", player.ID.Hex(), score)
	}
}

func UpdateAllPlayersPlace(context *util.Context) {
	var players []*Player
	context.DB.C(PlayerCollectionName).Find(nil).All(&players)

	for _, player := range players {
		player.UpdatePlace(context)
	}
}

func (player *Player) SetDirty(flags ...util.Bits) {
	for _, flag := range flags {
		player.DirtyMask = util.SetMask(player.DirtyMask, flag);
	}
}

func (player *Player) SetAllDirty() {
	player.DirtyMask = PlayerDataMask_All;
}

func (player *Player) MarshalDirty(context *util.Context) *map[string]interface{} {
	// check mask
	dirtyMask := player.DirtyMask
	if dirtyMask == PlayerDataMask_None {
		return nil
	}

	// create player data map
	dataMap := map[string]interface{} {}

	// check all updated data
	if util.CheckMask(dirtyMask, PlayerDataMask_Name) {
		dataMap["name"] = player.Name
		dataMap["tag"] = player.Tag
	}

	if util.CheckMask(dirtyMask, PlayerDataMask_Currency) {
		dataMap["standardCurrency"] = player.StandardCurrency
		dataMap["premiumCurrency"] = player.PremiumCurrency
	}

	if util.CheckMask(dirtyMask, PlayerDataMask_XP) {
		dataMap["level"] = player.GetLevel()
		dataMap["xp"] = player.XP
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Cards) {
		dataMap["cards"] = player.Cards
		dataMap["uncollectedCards"] = player.UncollectedCards
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Deck) {
		dataMap["decks"] = player.Decks
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Loadout) {
		dataMap["currentDeck"] = player.CurrentDeck
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Tomes) {
		dataMap["tomes"] = player.Tomes
		dataMap["arenaPoints"] = player.ArenaPoints
		dataMap["freeTomes"] = player.FreeTomes
		dataMap["freeTomeUnlockTime"] = player.FreeTomeUnlockTime
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Stars) {
		dataMap["rankPoints"] = player.RankPoints
		dataMap["rating"] = player.Rating
		dataMap["winCount"] = player.WinCount
		dataMap["lossCount"] = player.LossCount
		dataMap["matchCount"] = player.MatchCount
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Quests) {
		dataMap["quests"] = player.Quests
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Friends) {
		var friends []PlayerClient
		dataMap["friends"] = friends
	}
	
	if util.CheckMask(dirtyMask, PlayerDataMask_Guild) {
		if player.GuildID.Valid() {
			guild, err := GetGuildById(context, player.GuildID)
			if err != nil {
				log.Error(err)
			} else {
				dataMap["guild"], err = guild.CreateGuildClient(context)
				if err != nil {
					log.Error(err)
				}
			}

			dataMap["guildRole"] = player.GuildRole
		}
	}

	return &dataMap
}
