package models

import (
	"time"
	"fmt"
	"math/rand"
	"encoding/json"
	"io/ioutil"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
	"bloodtales/util"
	"bloodtales/log"
)

const PlayerCollectionName = "players"
const MinutesToUnlockFreeTome = 15

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
	Name                string          `bson:"-" json:"name"`
	XP                  int             `bson:"xp" json:"xp"`
	RankPoints          int             `bson:"rk" json:"rankPoints"`
	Rating              int             `bson:"rt" json:"rating"`

	WinCount            int             `bson:"wc" json:"winCount"`
	LossCount           int             `bson:"lc" json:"lossCount"`
	MatchCount          int             `bson:"mc" json:"matchCount"`

	StandardCurrency    int             `bson:"cs" json:"standardCurrency"`
	PremiumCurrency     int             `bson:"cp" json:"premiumCurrency"`
	Cards               []Card          `bson:"cd" json:"cards"`
	Decks               []Deck          `bson:"ds" json:"decks"`
	CurrentDeck         int             `bson:"dc" json:"currentDeck"`
	Tomes               []Tome          `bson:"tm" json:"tomes"`
	ArenaPoints         int             `bson:"ap" json:"arenaPoints"`
	FreeTomes           int             `bson:"ft" json:"freeTomes"`
	FreeTomeUnlockTime  int64           `bson:"fu" json:"freeTomeUnlockTime"`

	Quests              string          `bson:"qu,omitempty" json:"quests,omitempty"` // FIXME - temp fix until full quest system built on server

	FriendIDs           []bson.ObjectId `bson:"fd,omitempty" json:"-"`
	GuildID             bson.ObjectId   `bson:"gd,omitempty" json:"-"`

	DirtyMask           util.Bits       `bson:"-" json:"-"`

	CardsPurchased		[3]int 		    `bson:"pu" json:"-"`
	PurchaseResetTime 	int64 		    `bson:"pr" json:"-"`
}

// client model
type PlayerClient struct {
	Name                string   `json:"name"`
	Tag                 string   `json:"tag"`
	XP                  int      `json:"xp"`
	RankPoints          int      `json:"rankPoints"`
	Rating              int      `json:"rating"`
}

func ensureIndexPlayer(database *mgo.Database) {
	c := database.C(PlayerCollectionName)

	// username index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "us" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}))
}

func GetPlayerById(database *mgo.Database, id bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = database.C(PlayerCollectionName).Find(bson.M { "_id": id } ).One(&player)
	return
}

func GetPlayerByUser(database *mgo.Database, userId bson.ObjectId) (player *Player, err error) {
	// find player data by user ID
	err = database.C(PlayerCollectionName).Find(bson.M { "us": userId } ).One(&player)
	return
}

func (player *Player) initialize() {
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
	player.initialize()
	
	player.UserID = userID
	return
}

func (player *Player) CreatePlayerClient(database *mgo.Database) (client *PlayerClient, err error) {
	playerUser, err := GetUserById(database, player.UserID)
	if err != nil {
		return
	}

	client = &PlayerClient {
		Name: playerUser.Name,
		Tag: playerUser.Tag,
		XP: player.XP,
		RankPoints: player.RankPoints,
		Rating: player.Rating,
	}
	return
}

func (player *Player) Reset(database *mgo.Database) (err error) {
	// reset player and update in database
	player.initialize()
	return player.Save(database)
}

func UpdatePlayer(database *mgo.Database, user *User, data string) (player *Player, err error) {
	// find existing player data
	player, _ = GetPlayerByUser(database, user.ID)
	
	// initialize new player if none exists
	if player == nil {
		player = CreatePlayer(user.ID)
	}
	
	// parse updated data
	err = json.Unmarshal([]byte(data), &player)
	if err == nil {
		// update database
		err = player.Save(database)
	}
	return
}

func (player *Player) Save(database *mgo.Database) (err error) {
	if !player.ID.Valid() {
		player.ID = bson.NewObjectId()
	}

	// update entire player to database
	_, err = database.C(PlayerCollectionName).Upsert(bson.M { "us": player.UserID }, player)
	return
}

func (player *Player) GetLevel() int {
	return data.GetAccountLevel(player.XP)
}

func (player *Player) AddVictoryTome(database *mgo.Database) (tome *Tome, err error) {
	//first check to see if the player has an available tome slot, else return
	tome = nil
	for _, tomeSlot := range player.Tomes {
		if tomeSlot.State == TomeEmpty {
			tome = &tomeSlot
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
			(*tome).UnlockTime = data.TicksToTime(0)
			break
		}
	}

	err = player.Save(database)
	return
}

func (player *Player) AddRewards(database *mgo.Database, tome *Tome) (reward *TomeReward, err error) {
	reward = tome.OpenTome(player.GetLevel())
	player.PremiumCurrency += reward.PremiumCurrency
	player.StandardCurrency += reward.StandardCurrency

	for i, id := range reward.Cards {
		player.AddCards(id, reward.NumRewarded[i])
	}

	err = player.Save(database)
	return
}

func (player *Player) UpdateRewards(database *mgo.Database) error {
	unlockTime := data.TicksToTime(player.FreeTomeUnlockTime)

	for time.Now().UTC().After(unlockTime) && player.FreeTomes < 3 {
		unlockTime = unlockTime.Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute)
		player.FreeTomes++
	}

	player.FreeTomeUnlockTime = data.TimeToTicks(unlockTime)

	for i,_ := range player.Tomes {
		(&player.Tomes[i]).UpdateTome()
	}

	return player.Save(database)
}

func (player *Player) ClaimTome(database *mgo.Database, tomeId string) (*TomeReward, error) {
	tome := &Tome {
		DataID: data.ToDataId(tomeId),
	}

	// check currency
	// TODO

	return player.AddRewards(database, tome)
}

func (player *Player) ClaimFreeTome(database *mgo.Database) (tomeReward *TomeReward, err error) {
	err = player.UpdateRewards(database)

	if player.FreeTomes == 0 || err != nil {
		return
	}

	if player.FreeTomes == 3 {
		player.FreeTomeUnlockTime = data.TimeToTicks(time.Now().Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute))
	}

	player.FreeTomes--

	return player.ClaimTome(database, "TOME_COMMON")
}

func (player *Player) ClaimArenaTome(database *mgo.Database) (tomeReward *TomeReward, err error) {
	if player.ArenaPoints < 10 {
		return
	}

	player.ArenaPoints = 0;

	return player.ClaimTome(database, "TOME_RARE")
}

func (player *Player) ModifyArenaPoints(val int) {
	if val < 1 {
		return
	}

	player.ArenaPoints += val

	if player.ArenaPoints > 10 {
		player.ArenaPoints = 10
	}
}

func (player *Player) Delete(database *mgo.Database) (err error) {
	// delete player from database
	return database.C(PlayerCollectionName).Remove(bson.M { "_id": player.ID })
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

func (player *Player) HasCard(id data.DataId) (*Card, bool) {
	for _, card := range player.Cards {
		if card.DataID == id {
			return &card, true
		}
	}

	return nil, false
}

func (player *Player) AddCards(id data.DataId, num int) {
	//update the card if we already have it, otherwise instantiate a new one and add it in
	for i, card := range player.Cards {
		if card.DataID == id {
			player.Cards[i].CardCount += num
			return
		}
	}

	card := Card {
		DataID: id,
		Level: 1,
		CardCount: num,
		WinCount: 0,
		LeaderWinCount: 0,
	}

	player.Cards = append(player.Cards, card)
}

func (player *Player) GetMapOfCardIndexes() map[data.DataId]int {
	cardMap := map[data.DataId]int {}
	for index, card := range player.Cards {
		cardMap[card.DataID] = index
	}
	return cardMap
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
			guild, err := GetGuildById(context.DB, player.GuildID)
			if err != nil {
				log.Error(err)
			} else {
				dataMap["guild"], err = guild.CreateGuildClient(context.DB)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}

	return &dataMap
}
