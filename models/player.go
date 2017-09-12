package models

import (
	"fmt"
	"time"
	"io/ioutil"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/config"
	"bloodtales/data"
	"bloodtales/log"
	"bloodtales/util"
)

const PlayerCollectionName = "players"

const (
	PlayerDataMask_None     util.Bits 		= 0x0
	PlayerDataMask_All                		= 0xfffffff
	PlayerDataMask_Name               		= 0x1
	PlayerDataMask_Currency           		= 0x2
	PlayerDataMask_XP                 		= 0x4
	PlayerDataMask_Cards              		= 0x8
	PlayerDataMask_Deck               		= 0x10
	PlayerDataMask_Loadout            		= 0x20
	PlayerDataMask_Tomes              		= 0x40
	PlayerDataMask_Stars              		= 0x80
	PlayerDataMask_Quests             		= 0x100
	PlayerDataMask_Friends            		= 0x200
	PlayerDataMask_Guild              		= 0x400
	PlayerDataMask_Tutorial          		= 0x800
)

type Player struct {
	ID         				bson.ObjectId 	`bson:"_id,omitempty" json:"-"`
	UserID     				bson.ObjectId 	`bson:"us" json:"-"`
	TimeZone 				string 			`bson:"-" json:"-"`
	LastTime   				time.Time     	`bson:"tz" json:"-"`
	Name       				string        	`bson:"-" json:"name"`
	Tag        				string        	`bson:"-" json:"tag"`
	XP         				int           	`bson:"xp" json:"xp"`
	RankPoints 				int           	`bson:"rk" json:"rankPoints"`
	Rating     				int           	`bson:"rt" json:"rating"`

	WinCount   				int 			`bson:"wc" json:"winCount"`
	LossCount  				int 			`bson:"lc" json:"lossCount"`
	MatchCount 				int 			`bson:"mc" json:"matchCount"`

	StandardCurrency   		int    			`bson:"cs" json:"standardCurrency"`
	PremiumCurrency    		int    			`bson:"cp" json:"premiumCurrency"`
	Cards              		[]Card 			`bson:"cd" json:"cards"`
	UncollectedCards   		[]data.DataId	`bson:"uc" json:"uncollectedCards"`
	Decks              		[]Deck 			`bson:"ds" json:"decks"`
	CurrentDeck        		int    			`bson:"dc" json:"currentDeck"`
	Tomes              		[]Tome 			`bson:"tm" json:"tomes"`
	ActiveTome 				Tome 			`bson:"at" json:"activeTome"`
	ArenaPoints        		int    			`bson:"ap" json:"arenaPoints"`
	ArenaTomeUnlockTime 	int64 			`bson:"au" json:"arenaTomeUnlockTime"`
	FreeTomes          		int    			`bson:"ft" json:"freeTomes"`
	FreeTomeUnlockTime 		int64  			`bson:"fu" json:"freeTomeUnlockTime"`
	GuildTomeUnlockTime 	int64 			`bson:"gu" json:"guildTomeUnlockTime"`
	VictoryTomeCount 		int 			`bson:"vt"`

	Quests     				[]Quest 		`bson:"qu" json:"quests"`
	QuestClearTime 			int64       	`bson:"qc" json:"questClearTime"`
	

	GuildID   				bson.ObjectId 	`bson:"gd,omitempty" json:"-"`
	GuildRole 				GuildRole     	`bson:"gr,omitempty" json:"-"`
	GuildJoinTime			time.Time		`bson:"gj" json:"-"`

	TutorialDisabled		bool			`bson:"td" json:"tutorialDisabled"`
	Tutorial 				[]Tutorial 		`bson:"tl" json:"tutorial"`

	DirtyMask 				util.Bits 		`bson:"-" json:"-"`

	Store 					StoreHistory 	`bson:"sh"`
}

// client model
type PlayerClient struct {
	Name       				string 			`json:"name"`
	Tag        				string 			`json:"tag"`
	XP         				int    			`json:"xp"`
	RankPoints 				int    			`json:"rankPoints"`
	Rating     				int    			`json:"rating"`

	WinCount   				int 			`json:"winCount"`
	LossCount  				int 			`json:"lossCount"`
	MatchCount 				int 			`json:"matchCount"`

	GuildTag                string          `json:"guildTag"`
	GuildRole 				string  		`json:"guildRole"`
	GuildJoinTime			time.Time		`json:"guildJoinTime"`

	Online     				bool  			`json:"online"`
	LastOnline 				int64 			`json:"lastOnline"`
}

func ensureIndexPlayer(database *mgo.Database) {
	c := database.C(PlayerCollectionName)

	// username index
	util.Must(c.EnsureIndex(mgo.Index{
		Key:        []string{"us"},
		Unique:     true,
		DropDups:   true,
		Background: true,
	}))
}

func GetPlayerById(context *util.Context, id bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = context.DB.C(PlayerCollectionName).Find(bson.M{"_id": id}).One(&player)
	return
}

func GetPlayerByUser(context *util.Context, userId bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = context.DB.C(PlayerCollectionName).Find(bson.M{"us": userId}).One(&player)
	return
}

func GetPlayerByTag(context *util.Context, tag string) (player *Player, err error) {
	var user *User
	if user, err = GetUserByTag(context, tag); err != nil {
		return 
	}

	player, err = GetPlayerByUser(context, user.ID)
	return
}

func (player *Player) loadDefaults(development bool) (err error) {
	// template file path for initial player data
	path := "./resources/models/player.json"
	if development {
		path = "./resources/models/player-development.json"
	}

	// read template file
	var file []byte
	file, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	// setup tomes
	player.SetupTomeDefaults()

	// setup store data
	player.InitStore()

	err = json.Unmarshal(file, player)

	// assign starting quests (must happen after default cards are assigned - quests use player's card list)
	player.SetupQuestDefaults()

	return
}

func CreatePlayer(userID bson.ObjectId, development bool) (player *Player, err error) {
	player = &Player{}

	// load initial player data values
	err = player.loadDefaults(development)
	if err != nil {
		return
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
	online := (lastOnline < time.Second*config.Config.Sessions.OfflineTimeout)

	// get guild
	guildTag := ""
	guildRole := "None"
	if player.GuildID.Valid() {
		var guild *Guild
		guild, err = GetGuildById(context, player.GuildID)
		if err != nil {
			return
		}

		// guild name
		guildTag = guild.Tag

		// guild role
		guildRole = GetGuildRoleName(player.GuildRole)
	}

	// create player client
	client = &PlayerClient{
		Name:       playerUser.Name,
		Tag:        playerUser.Tag,
		XP:         player.XP,
		RankPoints: player.RankPoints,
		Rating:     player.Rating,

		WinCount:   player.WinCount,
		LossCount:  player.LossCount,
		MatchCount: player.MatchCount,

		GuildTag:   guildTag,
		GuildRole:  guildRole,

		Online:     online,
		LastOnline: util.DurationToTicks(lastOnline),
	}
	return
}

func (player *Player) Reset(context *util.Context, development bool) (err error) {
	// reset player data values
	err = player.loadDefaults(development)
	if err != nil {
		return
	}

	// clear cache for player
	playerID := player.ID.Hex()
	userID := player.UserID.Hex()
	context.Cache.Set(fmt.Sprintf("PlayerUserId:%s", playerID), nil)
	context.Cache.Set(fmt.Sprintf("PlayerName:%s", playerID), nil)
	context.Cache.Set(fmt.Sprintf("UserName:%s", userID), nil)
	context.Cache.RemoveScore("Leaderboard", playerID)

	player.GuildID = bson.ObjectId("")

	// update database
	return player.Save(context)
}

func ResetPlayers(context *util.Context) error {
	// TODO - this is an example of a bulk aggregate operation, but isn't fully tested...
	var result bson.D
	return context.DB.Run(bson.D{
		bson.DocElem{"update", PlayerCollectionName},
		bson.DocElem{"updates", []bson.M{
			bson.M{
				"q": bson.M{},
				"u": bson.M{
					"xp": 0,
					"rk": 0,
					"rt": 1200,
					"wc": 0,
					"lc": 0,
					"mc": 0,
					"ap": 0,
				},
				"multi":  false,
				"upsert": false,
				"limit":  0,
			},
		}},
		bson.DocElem{"writeConcern", bson.M{
			"w":        1,
			"j":        true,
			"wtimeout": 1000,
		}},
		bson.DocElem{"ordered", false},
	}, &result)
}

func (player *Player) Save(context *util.Context) (err error) {
	if !player.ID.Valid() {
		player.ID = bson.NewObjectId()
	}

	// last active time
	player.LastTime = time.Now()

	// update entire player to database
	_, err = context.DB.C(PlayerCollectionName).Upsert(bson.M{"_id": player.ID}, player)
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
	err = context.DB.C(PlayerCollectionName).Update(bson.M{"_id": player.ID}, bson.M{"$set": updates})
	return
}

func (player *Player) Delete(context *util.Context) (err error) {
	// delete player from database
	return context.DB.C(PlayerCollectionName).Remove(bson.M{"_id": player.ID})
}

func GetUserIdByPlayerId(context *util.Context, playerID bson.ObjectId) bson.ObjectId {
	// get cache key
	key := fmt.Sprintf("PlayerUserId:%s", playerID.Hex())

	// get cached ID
	userIDHex := context.Cache.GetString(key, "")
	var userID bson.ObjectId

	if bson.IsObjectIdHex(userIDHex) {
		// user cached ID
		userID = bson.ObjectIdHex(userIDHex)
	} else {
		// get and cache ID
		player, _ := GetPlayerById(context, playerID)
		if player != nil {
			userID = player.UserID
			context.Cache.Set(key, userID.Hex())
		}
	}
	return userID
}

func (player *Player) CacheName(context *util.Context, name string) {
	// get cache keys
	userKey := fmt.Sprintf("UserName:%s", player.UserID.Hex())
	playerKey := fmt.Sprintf("PlayerName:%s", player.ID.Hex())

	// refresh cached names
	context.Cache.Set(userKey, name)
	context.Cache.Set(playerKey, name)
}

func GetUserName(context *util.Context, userID bson.ObjectId) string {
	// get cache key
	key := fmt.Sprintf("UserName:%s", userID.Hex())

	// get cached name
	name := context.Cache.GetString(key, "")

	// immediately cache latest name
	if name == "" {
		user, err := GetUserById(context, userID)
		if err == nil && user != nil {
			context.Cache.Set(key, user.Name)
			name = user.Name
		}
	}
	return name
}

func GetPlayerName(context *util.Context, playerID bson.ObjectId) string {
	// get cache key
	key := fmt.Sprintf("PlayerName:%s", playerID.Hex())

	// get cached name
	name := context.Cache.GetString(key, "")

	// immediately cache latest name
	if name == "" {
		player, _ := GetPlayerById(context, playerID)
		if player != nil {
			user, _ := GetUserById(context, player.UserID)
			if user != nil {
				context.Cache.Set(key, user.Name)
				name = user.Name
			}
		}
	}
	return name
}

func (player *Player) GetLevel() int {
	return data.GetAccountLevel(player.XP)
}

func (player *Player) GetCard(cardId data.DataId) *Card {
	for i, card := range player.Cards {
		if card.DataID == cardId {
			return &player.Cards[i]
		}
	}
	return nil
}

func (player *Player) GetDeckCards(deckIndex int) (cards []*Card) {
	deck := player.Decks[deckIndex]
	cards = make([]*Card, 9)

	cards[0] = player.GetCard(deck.LeaderCardID)
	for i, cardId := range deck.CardIDs {
		cards[i + 1] = player.GetCard(cardId)
	}

	return
}

func (player *Player) AddVictoryTome(context *util.Context) (index int, tome *Tome) {
	index, tome = player.GetEmptyTomeSlot()

	if tome != nil {
		tome.DataID = data.GetNextVictoryTomeID(player.VictoryTomeCount)
		tome.State = TomeLocked
		tome.UnlockTime = 0
		tome.League = data.GetLeague(data.GetRank(player.RankPoints).Level)

		player.VictoryTomeCount++
	}

	return index, tome
}

func (player *Player) GetLeague() data.League {
	return data.GetLeague(data.GetRank(player.RankPoints).Level)
}

func (player *Player) GetDrawCount() int {
	return player.MatchCount - player.WinCount - player.LossCount
}

func (player *Player) GetWinRatio() string {
	if player.MatchCount > 0 {
		return fmt.Sprintf("%d%%", player.WinCount*100/player.MatchCount)
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
		rankInTier := rank.Level - (tier-1)*5
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
		winsFactor := (player.WinCount - player.LossCount) * 1000000 // / matches
		matchesFactor := matches * 1000
		pointsFactor := player.ArenaPoints

		score := winsFactor + matchesFactor + pointsFactor
		context.Cache.SetScore("Leaderboard", player.ID.Hex(), score)
	}
}

// HACK - inefficient
func UpdateAllPlayersPlace(context *util.Context) {
	// clear all previous places
	context.Cache.ClearScores("Leaderboard")

	var players []*Player
	context.DB.C(PlayerCollectionName).Find(nil).All(&players)

	for _, player := range players {
		player.UpdatePlace(context)
	}
}

func (player *Player) SetDirty(flags ...util.Bits) {
	for _, flag := range flags {
		player.DirtyMask = util.SetMask(player.DirtyMask, flag)
	}
}

func (player *Player) SetAllDirty() {
	player.DirtyMask = PlayerDataMask_All
}

func (player *Player) MarshalDirty(context *util.Context) *map[string]interface{} {
	// check mask
	dirtyMask := player.DirtyMask
	if dirtyMask == PlayerDataMask_None {
		return nil
	}

	// create player data map
	dataMap := map[string]interface{}{}

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

		uncollectedCards := make([]string, 0)
		for _, id := range player.UncollectedCards {
			uncollectedCards = append(uncollectedCards, data.ToDataName(id))
		}

		dataMap["uncollectedCards"] = uncollectedCards
	}

	if util.CheckMask(dirtyMask, PlayerDataMask_Deck) {
		dataMap["decks"] = player.Decks
	}

	if util.CheckMask(dirtyMask, PlayerDataMask_Loadout) {
		dataMap["currentDeck"] = player.CurrentDeck
	}

	if util.CheckMask(dirtyMask, PlayerDataMask_Tomes) {
		currentTime := util.TimeToTicks(time.Now().UTC())

		dataMap["tomes"] = player.Tomes
		dataMap["activeTome"] = &player.ActiveTome
		dataMap["arenaPoints"] = player.ArenaPoints
		dataMap["arenaTomeUnlockTime"] = player.ArenaTomeUnlockTime - currentTime
		dataMap["freeTomes"] = player.FreeTomes
		dataMap["freeTomeUnlockTime"] = player.FreeTomeUnlockTime - currentTime
		dataMap["guildTomeUnlockTime"] = player.GuildTomeUnlockTime - currentTime
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
		dataMap["questClearTime"] = player.QuestClearTime
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

			dataMap["guildRole"] = GetGuildRoleName(player.GuildRole)
		}
	}

	if util.CheckMask(dirtyMask, PlayerDataMask_Tutorial) {
		dataMap["tutorialDisabled"] = player.TutorialDisabled
		dataMap["tutorial"] = player.Tutorial
	}

	return &dataMap
}
