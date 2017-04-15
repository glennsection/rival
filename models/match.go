package models

import (
	"time"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const MatchCollectionName = "players"

// match type
type MatchType int
const (
	MatchUnranked MatchType = iota
	MatchRanked
	MatchElite
	MatchTournament
)

// match outcome
type MatchOutcome int
const (
	MatchLoss MatchOutcome = -1
	MatchDraw = 0
	MatchWin = 1
)

// server model
type Match struct {
	PlayerID      	bson.ObjectId `bson:"id1" json:"playerId"`
	OpponentID      bson.ObjectId `bson:"id2" json:"opponentId"`
	Type            MatchType     `bson:"tp" json:"type"`
	Outcome        	MatchOutcome  `bson:"oc" json:"outcome"`
	Time			time.Time     `bson:"ti" json:"time"`
}

// client model
type MatchClientAlias Match
type MatchClient struct {
	PlayerID      	string        `bson:"id1" json:"playerId"`
	OpponentID      string        `bson:"id2" json:"opponentId"`

	*MatchClientAlias
}

// custom marshalling
func (match *Match) MarshalJSON() ([]byte, error) {
	// create client model
	client := &MatchClient {
		PlayerID: match.PlayerID.Hex(),
		OpponentID: match.OpponentID.Hex(),
		MatchClientAlias: (*MatchClientAlias)(match),
	}
	
	// marshal with client model
	return json.Marshal(client)
}

// custom unmarshalling
func (match *Match) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &MatchClient {
		MatchClientAlias: (*MatchClientAlias)(match),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server player IDs
	match.PlayerID = bson.ObjectId(client.PlayerID)
	match.OpponentID = bson.ObjectId(client.OpponentID)

	return nil
}

func FindOpponentMatch(database *mgo.Database, match *Match) (opponentMatch *Match, err error) {
	// find opponent match
	err = database.C(MatchCollectionName).Find(bson.M {
		"id2": match.OpponentID,
		"time": bson.M {
			"$gt": match.Time.Add(-time.Minute),
			"$lt": match.Time.Add(time.Minute),
		},
	}).One(&opponentMatch)
	return
}

func (match *Match) GetPlayer(database *mgo.Database) (player *Player, err error) {
	return GetPlayerById(database, match.PlayerID)
}

func (match *Match) GetOpponent(database *mgo.Database) (player *Player, err error) {
	return GetPlayerById(database, match.OpponentID)
}
