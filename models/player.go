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
)

const PlayerCollectionName = "players"
const MinutesToUnlockFreeTome = 240

const (
	UpdateMask_None int64 = 0x0
	UpdateMask_Name = 0x1
	UpdateMask_Currency = 0x2
	UpdateMask_XP = 0x4
	UpdateMask_Cards = 0x8
	UpdateMask_Deck = 0x10
	UpdateMask_Loadout = 0x20
	UpdateMask_Tomes = 0x40
	UpdateMask_Stars = 0x80
    UpdateMask_Quests = 0x100
)

type Player struct {
	ID              		bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID         	 		bson.ObjectId `bson:"us" json:"-"`
	Name                	string        `bson:"-" json:"name"`
	XP 						int 		  `bson:"xp" json:"xp"`
	RankPoints          	int           `bson:"rk" json:"rankPoints"`
	Rating          		int           `bson:"rt" json:"rating"`

	WinCount       			int           `bson:"wc" json:"winCount"`
	LossCount       		int           `bson:"lc" json:"lossCount"`
	MatchCount       		int           `bson:"mc" json:"matchCount"`

	StandardCurrency 		int           `bson:"cs" json:"standardCurrency"`
	PremiumCurrency 		int           `bson:"cp" json:"premiumCurrency"`
	Cards           		[]Card        `bson:"cd" json:"cards"`
	Decks           		[]Deck        `bson:"ds" json:"decks"`
	CurrentDeck      		int           `bson:"dc" json:"currentDeck"`
	Tomes           		[]Tome        `bson:"tm" json:"tomes"`
	ArenaPoints		 		int 		  `bson:"ap" json:"arenaPoints"`
	FreeTomes		 		int 		  `bson:"ft" json:"freeTomes"`
	FreeTomeUnlockTime  	int64 		  `bson:"fu" json:"freeTomeUnlockTime"`

	Quests              	string        `bson:"qu,omitempty" json:"quests,omitempty"` // FIXME - temp fix until full quest system built on server

	GuildID             	bson.ObjectId `bson:"gd,omitempty" json:"-"`

	CardsPurchased			[3]int 		  `bson:"pu" json:"-"`
	PurchaseResetTime 		int64 		  `bson:"pr" json:"-"`
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

func (player *Player) HandleUpdateMask(updateMask int64, dataMap *map[string]interface{}) {
	if updateMask == UpdateMask_None {
		return
	}

	if (updateMask & UpdateMask_Name) == UpdateMask_Name {
		(*dataMap)["name"] = player.Name
	}

	if (updateMask & UpdateMask_Currency) == UpdateMask_Currency {
		(*dataMap)["standardCurrency"] = player.StandardCurrency
		(*dataMap)["premiumCurrency"] = player.PremiumCurrency
	}

	if (updateMask & UpdateMask_XP) == UpdateMask_XP {
		(*dataMap)["level"] = player.GetLevel()
		(*dataMap)["xp"] = player.XP
	}
	
	if (updateMask & UpdateMask_Cards) == UpdateMask_Cards {
		(*dataMap)["cards"] = player.Cards
	}
	
	if (updateMask & UpdateMask_Deck) == UpdateMask_Deck {
		(*dataMap)["decks"] = player.Decks
	}
	
	if (updateMask & UpdateMask_Loadout) == UpdateMask_Loadout {
		(*dataMap)["currentDeck"] = player.CurrentDeck
	}
	
	if (updateMask & UpdateMask_Tomes) == UpdateMask_Tomes {
		(*dataMap)["tomes"] = player.Tomes
		(*dataMap)["arenaPoints"] = player.ArenaPoints
		(*dataMap)["freeTomes"] = player.FreeTomes
		(*dataMap)["freeTomeUnlockTime"] = player.FreeTomeUnlockTime
	}
	
	if (updateMask & UpdateMask_Stars) == UpdateMask_Stars {
		(*dataMap)["rankPoints"] = player.RankPoints
		(*dataMap)["rating"] = player.Rating
		(*dataMap)["winCount"] = player.WinCount
		(*dataMap)["lossCount"] = player.LossCount
		(*dataMap)["matchCount"] = player.MatchCount
	}
	
    if (updateMask & UpdateMask_Quests) == UpdateMask_Quests {
    	(*dataMap)["quests"] = player.Quests
	}
}
