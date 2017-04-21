package models

import (
	"time"
	"fmt"
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
)

const PlayerCollectionName = "players"
const MinutesToUnlockFreeTome = 240

type Player struct {
	ID              	bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID         	 	bson.ObjectId `bson:"us" json:"-"`
	Name            	string        `bson:"nm" json:"name"`
	Level           	int           `bson:"lv" json:"level"`
	Rank            	int           `bson:"rk" json:"rank"`
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

func UpdatePlayer(database *mgo.Database, user *User, data string) (err error) {
	// find existing player data
	player, _ := GetPlayerByUser(database, user.ID)
	
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

func (player *Player) AddRewards(database *mgo.Database, tome *Tome) (reward *TomeReward, err error) {
	reward = tome.OpenTome(player.Level)
	player.PremiumCurrency += reward.PremiumCurrency
	player.StandardCurrency += reward.StandardCurrency

	// since we can have up to 6 cards rewarded and they're unsorted, it can take up to O(6n) to see if our card
	// list already contains the cards we want to add. if we instead create a map of indexes to cards, we incur a 
	// cost of O(n) to create the map, and then have O(1) access time thereafter at the cost of memory
	cardMap := map[data.DataId]int {}
	for index, card := range player.Cards {
		cardMap[card.DataID] = index
	}

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
	currentTime := data.TimeToTicks(time.Now())
	unlockTime := data.TicksToTime(player.FreeTomeUnlockTime)

	for currentTime > player.FreeTomeUnlockTime && player.FreeTomes < 3 {
		unlockTime = unlockTime.Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute)
		player.FreeTomeUnlockTime = data.TimeToTicks(unlockTime)
		player.FreeTomes++
	}

	for i,_ := range player.Tomes {
		(&player.Tomes[i]).UpdateTome()
	}

	err = player.Update(database)

	return
}

func (player *Player) ClaimFreeReward(database *mgo.Database) (reward *TomeReward, err error) {
	err = player.UpdateRewards(database)

	if player.FreeTomes == 0 || err != nil {
		return
	}

	if(player.FreeTomes == 3) {
		player.FreeTomeUnlockTime = data.TimeToTicks(time.Now().Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute))
	}

	player.FreeTomes--
	tome := &Tome {
		DataID: data.ToDataId("TOME_COMMON"),
	}

	reward, err = player.AddRewards(database, tome) 
	if err != nil {
		panic(err)
	}

	return
}

func (player *Player) Delete(database *mgo.Database) (err error) {
	// delete player from database
	return database.C(PlayerCollectionName).Remove(bson.M { "_id": player.ID })
}

func (player *Player) GetRankData() *data.RankData {
	return data.GetRank(player.Rank)
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
