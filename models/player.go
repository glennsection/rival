package models

import (
	"time"
	"fmt"
	"math/rand"
	"encoding/json"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
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
	ID              	bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID         	 	bson.ObjectId `bson:"us" json:"-"`
	Name            	string        `bson:"nm" json:"name"`
	Level           	int           `bson:"lv" json:"level"`
	Xp 					int 		  `bson:"xp" json:"xp"`
	RankPoints          int           `bson:"rk" json:"rankPoints"`
	Rating          	int           `bson:"rt" json:"rating"`
	WinCount       		int           `bson:"wc" json:"winCount"`
	LossCount       	int           `bson:"lc" json:"lossCount"`
	MatchCount       	int           `bson:"mc" json:"matchCount"`

	StandardCurrency 	int           `bson:"cs" json:"standardCurrency"`
	PremiumCurrency 	int           `bson:"cp" json:"premiumCurrency"`
	Cards           	[]Card        `bson:"cd" json:"cards"`
	Decks           	[]Deck        `bson:"ds" json:"decks"`
	CurrentDeck      	int           `bson:"dc" json:"currentDeck"`
	Tomes           	[]Tome        `bson:"tm" json:"tomes"`
	ArenaPoints		 	int 		  `bson:"ap" json:"arenaPoints"`
	FreeTomes		 	int 		  `bson:"ft" json:"freeTomes"`
	FreeTomeUnlockTime  int64 		  `bson:"fu" json:"freeTomeUnlockTime"`

	Quests              string        `bson:"qu,omitempty" json:"quests,omitempty"` // FIXME - temp fix until full quest system built on server
}

func ensureIndexPlayer(database *mgo.Database) {
	c := database.C(PlayerCollectionName)

	index := mgo.Index {
		Key:        []string { "us" },
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
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

func CreatePlayer(userID bson.ObjectId, name string) (player *Player) {
	player = &Player {}
	player.Initialize()
	player.ID = bson.NewObjectId()
	player.UserID = userID
	player.Name = name
	return
}

func (player *Player) Reset(database *mgo.Database) (err error) {
	// reset player and update in database
	player.Initialize()
	return player.Update(database)
}

func UpdatePlayer(database *mgo.Database, user *User, data string) (player *Player, err error) {
	// find existing player data
	player, _ = GetPlayerByUser(database, user.ID)
	
	// initialize new player if none exists
	if player == nil {
		player = CreatePlayer(user.ID, user.Username)
	}
	
	// parse updated data
	err = json.Unmarshal([]byte(data), &player)
	if err == nil {
		// update database
		err = player.Update(database)
	}
	return
}

func (player *Player) Update(database *mgo.Database) (err error) {
	// update entire player to database
	_, err = database.C(PlayerCollectionName).Upsert(bson.M { "us": player.UserID }, player)
	return
}

func (player *Player) AddVictoryTome(database *mgo.Database) (tome *Tome) {
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

	player.Update(database)

	return
}

func (player *Player) AddRewards(database *mgo.Database, tome *Tome) (reward *TomeReward, err error) {
	reward = tome.OpenTome(data.GetAccountLevel(player.Xp))
	player.PremiumCurrency += reward.PremiumCurrency
	player.StandardCurrency += reward.StandardCurrency

	// since we can have up to 6 cards rewarded and they're unsorted, it can take up to O(6n) to see if our card
	// list already contains the cards we want to add. if we instead create a map of indexes to cards, we incur a 
	// cost of O(n) to create the map, and then have O(1) access time thereafter at the cost of memory
	cardMap := player.GetMapOfCardIndexes()

	for i, id := range reward.Cards {
		//update the card if we already have it, otherwise instantiate a new one and add it in
		if index, hasCard := cardMap[id]; hasCard {
			player.Cards[index].CardCount += reward.NumRewarded[i]
		} else {
			card := Card{
				DataID: id,
				Level: 1,
				CardCount: reward.NumRewarded[i],
				WinCount: 0,
				LeaderWinCount: 0,
			}

			player.Cards = append(player.Cards, card)
		}
	}

	err = player.Update(database)
	
	return
}

func (player *Player) UpdateRewards(database *mgo.Database) (err error){
	unlockTime := data.TicksToTime(player.FreeTomeUnlockTime)

	for time.Now().UTC().After(unlockTime) && player.FreeTomes < 3 {
		unlockTime = unlockTime.Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute)
		player.FreeTomes++
	}

	player.FreeTomeUnlockTime = data.TimeToTicks(unlockTime)

	for i,_ := range player.Tomes {
		(&player.Tomes[i]).UpdateTome()
	}

	err = player.Update(database)

	return
}

func (player *Player) ClaimTome(database *mgo.Database, tomeId string) (reward *TomeReward, err error) {
	tome := &Tome {
		DataID: data.ToDataId(tomeId),
	}

	// check currency
	// TODO

	reward, err = player.AddRewards(database, tome)

	return
}

func (player *Player) ClaimFreeTome(database *mgo.Database) (reward *TomeReward, err error) {
	err = player.UpdateRewards(database)

	if player.FreeTomes == 0 || err != nil {
		return
	}

	if player.FreeTomes == 3 {
		player.FreeTomeUnlockTime = data.TimeToTicks(time.Now().Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute))
	}

	player.FreeTomes--

	reward, err = player.ClaimTome(database, "TOME_COMMON")

	return
}

func (player *Player) ClaimArenaTome(database *mgo.Database) (reward *TomeReward, err error) {
	if player.ArenaPoints < 10 {
		return
	}

	player.ArenaPoints = 0;

	reward, err = player.ClaimTome(database, "TOME_RARE")

	return
}

func (player *Player) ModifyArenaPoints(val int) {
	if val < 1 {
		return
	}

	player.ArenaPoints += val

	if player.ArenaPoints > 10 {
		player.ArenaPoints = 10
	}

	return
}

func (player *Player) Delete(database *mgo.Database) (err error) {
	// delete player from database
	return database.C(PlayerCollectionName).Remove(bson.M { "_id": player.ID })
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
		return fmt.Sprintf("Tier %d Rank %d", tier, rankInTier)
	}
	return "Unranked"
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
		(*dataMap)["level"] = player.Level
		(*dataMap)["xp"] = player.Xp
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
