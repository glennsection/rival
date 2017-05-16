package models

import (
	"time"
	"math"
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
	MatchSurrender MatchOutcome = -2
	MatchLoss = -1
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
	PlayerScore		int 		  `bson:"ps" json:"playerScore"`
	OpponentScore 	int 		  `bson:"os" json:"opponentScore"`
	StartTime	    time.Time     `bson:"t0" json:"-"`
	EndTime	        time.Time     `bson:"t1" json:"-"`

	// client
	Host            bool          `bson:"-" json:"host"`

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

// client rewards
type MatchReward struct {
	Tome 			*Tome 		  `json:"tome"`
	ArenaPoints 	int 		  `json:"arenaPoints"`
}

func ensureIndexMatch(database *mgo.Database) {
	c := database.C(MatchCollectionName)

	// player index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "id1", "id2", "state" },
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}))
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

func GetMatchById(database *mgo.Database, id bson.ObjectId) (match *Match, err error) {
	err = database.C(MatchCollectionName).Find(bson.M { "_id": id } ).One(&match)
	return
}

func (match *Match) Save(database *mgo.Database) (err error) {
	if !match.ID.Valid() {
		match.ID = bson.NewObjectId()
	}

	// update match in database
	_, err = database.C(MatchCollectionName).Upsert(bson.M { "_id": match.ID }, match)
	return
}

func (match *Match) Delete(database *mgo.Database) (err error) {
	return database.C(MatchCollectionName).Remove(bson.M { "_id": match.ID })
}

func ClearMatches(database *mgo.Database, player *Player) (err error) {
	// find and remove all invalid matches with player
	_, err = database.C(MatchCollectionName).RemoveAll(bson.M {
		"$or": []bson.M {
			bson.M { "id1": player.ID, },
			bson.M { "id2": player.ID, },
		},
		"st": bson.M {
			"$in": []interface{} {
				MatchInvalid,
				MatchOpen,
				MatchActive,
			},
		},
 	})
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

	//log.Printf("FindMatch(%v [%v], %v) => %v", player.Name, player.ID, matchType, match)

	if match != nil {
		// match players and mark as active
		match.OpponentID = player.ID
		match.State = MatchActive
		match.StartTime = time.Now()
		match.Host = false
	} else {
		// queue new match
		match = &Match {
			PlayerID: player.ID,
			Type: matchType,
			RoomID: util.GenerateUUID(),
			State: MatchOpen,
			StartTime: time.Now(),
			Host: true,
		}
	}

	// update database
	err = match.Save(database)
	return
}

func FailMatch(database *mgo.Database, player *Player) (err error) {
	// find and remove all invalid matches with player
	var matches []*Match
	err = database.C(MatchCollectionName).Find(bson.M {
		"$or": []bson.M {
			bson.M { "id1": player.ID, },
			bson.M { "id2": player.ID, },
		},
		"st": bson.M {
			"$in": []interface{} {
				MatchOpen,
				MatchActive,
			},
		},
	}).All(&matches)

	if err == nil {
		// fix all found matches
		for _, match := range matches {
			if match.State == MatchActive {
				if match.PlayerID == player.ID {
					match.PlayerID = match.OpponentID
					match.OpponentID = bson.ObjectId("")
				} else {
					match.OpponentID = bson.ObjectId("")
				}
				match.State = MatchOpen
				match.Save(database)
			} else {
				match.Delete(database)
			}
		}
	}
	return
}

func CompleteMatch(database *mgo.Database, player *Player, host bool, outcome MatchOutcome, playerScore int, opponentScore int) (match *Match, matchReward *MatchReward, err error) {
	// prepare match change
	change := mgo.Change {
		Upsert: false,
		ReturnNew: true,
	}

	// check if host or guest
	var idField string
	if host {
		// prepare host query
		idField = "id1"

		// prepare host match change
		change.Update = bson.M {
			"$set": bson.M {
				"st": MatchCompleting,
				"oc": outcome,
				"ps": playerScore,
				"os": opponentScore,
				"t1": time.Now(),
			},
		}
	} else {
		// prepare guest query
		idField = "id2"

		// invert outcome for guest
		outcome = invertOutcome(outcome)

		// invert scores for guest
		temp := playerScore
		playerScore = opponentScore
		opponentScore = temp

		// prepare guest match change
		change.Update = bson.M {
			"$set": bson.M {
				"st": MatchCompleting,
				"oc": outcome,
				"ps": playerScore,
				"os": opponentScore,
				"t1": time.Now(),
			},
		}
	}

	// find active match for player, and update if found
	foundActiveMatch := true
	_, err = database.C(MatchCollectionName).Find(bson.M {
		idField: player.ID,
		"st": MatchActive,
	}).Apply(change, &match)
	if err != nil {
		if err.Error() == "not found" {
			// continue without error
			match = nil
			err = nil
			foundActiveMatch = false
		} else {
			err = util.NewError(err)
			return
		}
	}

	if foundActiveMatch {
		// update player stats
		err = match.ProcessMatchResults(database)
		if err != nil {
			err = util.NewError(err)
			return
		}
	} else {
		// prepare match change
		change = mgo.Change {
			Update: bson.M { "$set": bson.M { "st": MatchComplete }, },
			Upsert: false,
			ReturnNew: true,
		}

		// find completing match, and set to completed if found
		_, err = database.C(MatchCollectionName).Find(bson.M {
			idField: player.ID,
			"st": MatchCompleting,
		}).Apply(change, &match)
		if err != nil {
			err = util.NewError(err)
			return
		}
		// TODO - make sure we check if match was found

		// validate match outcome
		log.Printf("%v:%v %d:%d %d:%d", match.Outcome, outcome, match.PlayerScore, playerScore, match.OpponentScore, opponentScore)
		if (match.Outcome == outcome && match.PlayerScore == playerScore && match.OpponentScore == opponentScore) || match.Outcome == MatchSurrender || outcome == MatchSurrender {
			// match.State = MatchComplete
		} else {
			match.State = MatchInvalid

			// update as invalid
			match.Save(database)

			err = util.NewError("Non-symmetrical match outcomes reported by clients!")

			// TODO - roll back player stats!
		}
	}

	if match.State != MatchInvalid && err == nil && outcome != MatchSurrender {
		matchReward = &MatchReward {}

		if host {
			player.ModifyArenaPoints(match.PlayerScore)
			matchReward.ArenaPoints = match.PlayerScore
		} else {
			player.ModifyArenaPoints(match.OpponentScore)
			matchReward.ArenaPoints = match.OpponentScore
		}

		if (host && match.Outcome == MatchWin) || (!host && match.Outcome == MatchLoss) {
			matchReward.Tome = player.AddVictoryTome(database)
		} else {
			player.Save(database)
		}
	} 

	return
}

func (match *Match) GetPlayer(database *mgo.Database) (player *Player, err error) {
	err = nil
	if match.player == nil {
		if match.PlayerID.Valid() {
			match.player, err = GetPlayerById(database, match.PlayerID)
		}
	}
	return match.player, err
}

func (match *Match) GetOpponent(database *mgo.Database) (player *Player, err error) {
	err = nil
	if match.opponent == nil {
		if match.OpponentID.Valid() {
			match.opponent, err = GetPlayerById(database, match.OpponentID)
		}
	}
	return match.opponent, err
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
	case MatchSurrender:
		return MatchSurrender
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

		if (rankChange == -2) { // handle surrender - players should never lose more than one rank
			rankChange = -1
		}

		if rankChange > 0 || player.GetRankTier() > 1 {
			player.RankPoints += rankChange
		}
		if rankChange < 0 || opponent.GetRankTier() > 1 {
			opponent.RankPoints -= rankChange
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
	err = player.Save(database)
	if err != nil {
		return
	}
	err = opponent.Save(database)
	if err != nil {
		return
	}
	return
}
