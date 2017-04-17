package models

import (
	"time"
	"math"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
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

// match state
type MatchState int
const (
	MatchInvalid MatchState = iota
	MatchOpen
	MatchActive
	MatchComplete
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
	State           MatchState    `bson:"st" json:"state"`
	Outcome        	MatchOutcome  `bson:"oc" json:"outcome"`
	Time			time.Time     `bson:"ti" json:"time"`
}

// client model
type MatchClientAlias Match
type MatchClient struct {
	PlayerID      	string        `json:"playerId"`
	OpponentID      string        `json:"opponentId"`
	State           string        `json:"state"`

	*MatchClientAlias
}

// custom marshalling
func (match *Match) MarshalJSON() ([]byte, error) {
	// create client model
	client := &MatchClient {
		PlayerID: match.PlayerID.Hex(),
		OpponentID: match.OpponentID.Hex(),
		State: match.GetStateName(),
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
	match.State = parseStateName(client.State)

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

func (match *Match) GetStateName() string {
	switch match.State {
	default:
		return "Invalid"
	case MatchOpen:
		return "Open"
	case MatchActive:
		return "Active"
	case MatchComplete:
		return "Complete"
	}
}

func parseStateName(name string) MatchState {
	switch name {
	default:
		return MatchInvalid
	case "Open":
		return MatchOpen
	case "Active":
		return MatchActive
	case "Complete":
		return MatchComplete
	}
}

func getKFactor(playerRating int, opponentRating int) float64 {
	// chess k-factors  (TODO - work on this...)
	rating := util.Min(playerRating, opponentRating)
	if rating < 2100 {
		return 32.0
	} else if rating < 2400 {
		return 24.0
	}
	return 16.0
}

func (match *Match) ProcessMatchResults(database *mgo.Database) {
	// get players
	player, err := match.GetPlayer(database)
	if err != nil {
		panic(err)
	}
	opponent, err := match.GetOpponent(database)
	if err != nil {
		panic(err)
	}

	// update according to match type
	switch match.Type {

	case MatchRanked:
		// update stats
		player.Rank += int(match.Outcome)
		opponent.Rank -= int(match.Outcome)

	case MatchElite:
		// get k-factor
		kFactor := getKFactor(player.Rating, opponent.Rating)

		// transformed ratings
		q1 := math.Pow10(player.Rating / 400)
		q2 := math.Pow10(opponent.Rating / 400)
		qs := q1 + q2

		// expected scores
		e1 := q1 / qs
		e2 := q2 / qs

		// observed scores
		s1 := 0.5 + float64(match.Outcome) * 0.5
		s2 := 1 - s1

		// resulting ratings
		r1 := player.Rating + util.RoundToInt(kFactor * (s1 - e1))
		r2 := opponent.Rating + util.RoundToInt(kFactor * (s2 - e2))

		//log.Printf("Elite Match Results: [%v(%v) + %v:%v => %v] vs. [%v(%v) + %v:%v => %v]", player.Rating, q1, e1, s1, r1, opponent.Rating, q2, e2, s2, r2)
		
		// update stats
		player.Rating = r1
		opponent.Rating = r2

	case MatchTournament:
		// TODO

	}

	// modify win/loss counts and update database
	player.MatchCount += 1
	opponent.MatchCount += 1
	switch match.Outcome {

	case MatchWin:
		player.WinCount += 1
		opponent.LossCount += 1

	case MatchLoss:
		player.LossCount += 1
		opponent.WinCount += 1

	}
	player.Update(database)
	opponent.Update(database)
}