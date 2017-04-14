package models

import (
	"time"
	// "encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "bloodtales/data"
)

const MatchCollectionName = "players"

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
	Outcome        	MatchOutcome  `bson:"oc" json:"outcome"`
	Time			time.Time     `bson:"ti" json:"time"`
}

/* FIXME - this shouldn't be needed if Time is just set by the API call
// client model
type MatchClientAlias Match
type MatchClient struct {
	//Outcome        	string           `json:"outcome"`
	Time            int64         `json:"time"`

	*MatchClientAlias
}

// custom marshalling
func (match *Match) MarshalJSON() ([]byte, error) {
	// create client model
	client := &MatchClient {
		Time: data.TimeToTicks(match.Time),
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

	// server time
	match.Time = data.TicksToTime(client.Time)

	return nil
}
*/

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
