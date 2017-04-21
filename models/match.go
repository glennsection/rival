package models

import (
	"time"
	"math"
	"errors"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/log"
)

const MatchCollectionName = "matches"

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
	MatchCompleting
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
	ID              bson.ObjectId `bson:"_id,omitempty" json:"-"`
	PlayerID     	bson.ObjectId `bson:"id1" json:"-"`
	OpponentID      bson.ObjectId `bson:"id2,omitempty" json:"-"`
	Type            MatchType     `bson:"tp" json:"-"`
	RoomID          string        `bson:"rm" json:"roomId"`
	State           MatchState    `bson:"st" json:"state"`
	Outcome       	MatchOutcome  `bson:"oc" json:"outcome"`
	StartTime	    time.Time     `bson:"t0" json:"-"`
	EndTime	        time.Time     `bson:"t1" json:"-"`

	// internal
	player          *Player
	opponent        *Player
}

// client model
type MatchClientAlias Match
type MatchClient struct {
	State           string        `json:"state"`

	*MatchClientAlias
}

func ensureIndexMatch(database *mgo.Database) {
	c := database.C(MatchCollectionName)

	// player index
	index := mgo.Index {
		Key:        []string { "id1", "id2", "state" },
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

// custom marshalling
func (match *Match) MarshalJSON() ([]byte, error) {
	// create client model
	client := &MatchClient {
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

	// server state
	match.State = parseStateName(client.State)

	return nil
}

func (match *Match) Update(database *mgo.Database) (err error) {
	// update match in database
	_, err = database.C(MatchCollectionName).Upsert(bson.M { "id1": match.PlayerID }, match)
	return
}

func FindMatch(database *mgo.Database, player *Player, matchType MatchType) (match *Match, err error) {
	// find existing match (TODO - verify that no other pending matches exist for player)
	err = database.C(MatchCollectionName).Find(bson.M {
		"id1": bson.M {
			"$ne": player.ID,
		},
		"st": MatchOpen,
		"tp": matchType,
	}).One(&match)

	log.Printf("FindMatch(%v [%v], %v) => %v", player.Name, player.ID, matchType, match)

	if match != nil {
		// match players and mark as active
		match.OpponentID = player.ID
		match.State = MatchActive
		match.StartTime = time.Now()
	} else {
		// queue new match
		match = &Match {
			ID: bson.NewObjectId(),
			PlayerID: player.ID,
			Type: matchType,
			RoomID: util.GenerateUUID(),
			State: MatchOpen,
			StartTime: time.Now(),
		}
	}

	// update database
	err = match.Update(database)
	return
}

func CompleteMatch(database *mgo.Database, player *Player, outcome MatchOutcome) (err error) {
	// find active match for player
	var match *Match
	err = database.C(MatchCollectionName).Find(bson.M {
		"$or": []bson.M {
			bson.M { "id1": player.ID, },
			bson.M { "id2": player.ID, },
		},
		"st": bson.M {
			"$in": []interface{} {
				MatchActive,
				MatchCompleting,
			},
		},
 	}).One(&match)
 	if err != nil {
 		return
 	}

 	log.Printf("CompleteMatch(player: %v, match: %v)", player.ID, match)

 	// determine if player is match owner, and alter outcome accordingly
 	owner := (match.PlayerID == player.ID)
	if owner == false {
		outcome = invertOutcome(outcome)
	}

	if match.State == MatchActive {
		// update match outcome
		match.State = MatchCompleting
		match.Outcome = outcome
		match.EndTime = time.Now()

		// update database
		err = match.Update(database)
		if err != nil {
			return
		}

		// update player stats
		err = match.ProcessMatchResults(database)
	} else {
		// validate match outcome
		if match.Outcome == outcome {
			match.State = MatchComplete
		} else {
			match.State = MatchInvalid

			err = errors.New("Non-symmetrical match outcomes reported by clients!")

			// TODO - roll back stats!
		}

		// update as invalid
		match.Update(database)
	}
	return
}

func (match *Match) LoadPlayers(database *mgo.Database) (err error) {
	match.player, err = match.GetPlayer(database)
	if err != nil {
		return
	}

	match.opponent, err = match.GetOpponent(database)
	return
}

func (match *Match) GetPlayer(database *mgo.Database) (player *Player, err error) {
	if match.PlayerID.Valid() {
		return GetPlayerById(database, match.PlayerID)
	}
	return nil, nil
}

func (match *Match) GetPlayerName() string {
	if match.player != nil {
		return match.player.Name
	}
	return "None"
}

func (match *Match) GetOpponent(database *mgo.Database) (player *Player, err error) {
	if match.OpponentID.Valid() {
		return GetPlayerById(database, match.OpponentID)
	}
	return nil, nil
}

func (match *Match) GetOpponentName() string {
	if match.opponent != nil {
		return match.opponent.Name
	}
	return "None"
}

func (match *Match) GetTypeName() string {
	switch match.Type {
	default:
		return "Invalid"
	case MatchUnranked:
		return "Unranked"
	case MatchRanked:
		return "Ranked"
	case MatchElite:
		return "Elite"
	case MatchTournament:
		return "Tournament"
	}
}

func (match *Match) GetStateName() string {
	switch match.State {
	default:
		return "Invalid"
	case MatchOpen:
		return "Open"
	case MatchActive:
		return "Active"
	case MatchCompleting:
		return "Completing"
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
	case "Completing":
		return MatchCompleting
	case "Complete":
		return MatchComplete
	}
}

func (match *Match) GetOutcomeName() string {
	switch match.State {
	default:
		return "-"
	case MatchCompleting, MatchComplete:
		switch match.Outcome {
		default:
			return "Draw"
		case -1:
			return "Player 2 Win"
		case 1:
			return "Player 1 Win"
		}
	}
}

func invertOutcome(outcome MatchOutcome) MatchOutcome {
	switch outcome {
	case MatchLoss:
		return MatchWin
	case MatchWin:
		return MatchLoss
	default:
		return MatchDraw
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

func (match *Match) ProcessMatchResults(database *mgo.Database) (err error) {
	// get players
	player, err := match.GetPlayer(database)
	if err != nil {
		return
	}
	opponent, err := match.GetOpponent(database)
	if err != nil {
		return
	}

	// update according to match type
	switch match.Type {

	case MatchRanked:
		// update stats
		rankChange := int(match.Outcome)
		if rankChange > 0 || player.GetRankTier() > 1 {
			player.Rank += rankChange
		}
		if rankChange < 0 || opponent.GetRankTier() > 1 {
			opponent.Rank -= rankChange
		}

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
	err = player.Update(database)
	if err != nil {
		return
	}
	err = opponent.Update(database)
	return
}

func (match *Match) Delete(database *mgo.Database) (err error) {
	return database.C(MatchCollectionName).Remove(bson.M { "_id": match.ID })
}

func GetMatchById(database *mgo.Database, id bson.ObjectId) (match *Match, err error) {
	err = database.C(MatchCollectionName).Find(bson.M { "_id": id } ).One(&match)
	return
}
