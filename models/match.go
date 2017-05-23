package models

import (
	"fmt"
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
	HostID          bson.ObjectId `bson:"id1" json:"-"`
	OpponentID      bson.ObjectId `bson:"id2,omitempty" json:"-"`
	Type            MatchType     `bson:"tp" json:"-"`
	RoomID          string        `bson:"rm" json:"roomId"`
	State           MatchState    `bson:"st" json:"state"`
	Outcome       	MatchOutcome  `bson:"oc" json:"outcome"`
	HostScore		int 		  `bson:"s1" json:"hostScore"`
	OpponentScore 	int 		  `bson:"s2" json:"opponentScore"`
	StartTime	    time.Time     `bson:"t0" json:"-"`
	EndTime	        time.Time     `bson:"t1" json:"-"`

	// client
	Hosting         bool          `bson:"-" json:"hosting"`

	// internal
	host            *Player
	opponent        *Player
}

// client model
type MatchClientAlias Match
type MatchClient struct {
	State           string        `json:"state"`

	*MatchClientAlias
}

// cached match results
type MatchResult struct {
	MatchID         bson.ObjectId `json:"mid"`
	Outcome       	MatchOutcome  `json:"oc"`
	HostScore		int 		  `json:"s1"`
	OpponentScore 	int 		  `json:"s2"`
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
		Key:        []string { "rm" },
		Unique:     true,
		DropDups:   true,
		Background: true,
	}))

	// player index
	util.Must(c.EnsureIndex(mgo.Index {
		Key:        []string { "id1", "st", "tp" },
		Background: true,
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

func (matchResult *MatchResult) String() string {
	raw, err := json.Marshal(matchResult)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(raw)
}

func GetMatchResultByMatchId(context *util.Context, roomID string) (matchResult *MatchResult, ok bool) {
	// get cache key
	key := fmt.Sprintf("MatchResult:%s", roomID)

	// get cached result
	ok = context.Cache.GetJSON(key, &matchResult)
	return
}

func SetMatchResult(context *util.Context, roomID string, matchResult *MatchResult) {
	// get cache key
	key := fmt.Sprintf("MatchResult:%s", roomID)

	// get cached result
	context.Cache.Set(key, matchResult)
}

func ClearMatchResult(context *util.Context, roomID string) {
	SetMatchResult(context, roomID, nil)
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
		match.Hosting = false
	} else {
		// queue new match
		match = &Match {
			HostID: player.ID,
			Type: matchType,
			RoomID: util.GenerateUUID(),
			State: MatchOpen,
			StartTime: time.Now(),
			Hosting: true,
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
				if match.HostID == player.ID {
					match.HostID = match.OpponentID
				}
				match.OpponentID = bson.ObjectId("")
				match.State = MatchOpen
				match.Save(database)

				// TODO FIXME - need to send (via websocket), to new host, the fact that they are now the host
			} else {
				match.Delete(database)
			}
		}
	}
	return
}

func CompleteMatch(context *util.Context, player *Player, roomID string, outcome MatchOutcome, playerScore int, opponentScore int) (match *Match, matchReward *MatchReward, err error) {
	database := context.DB

	// get match from database
	err = database.C(MatchCollectionName).Find(bson.M {
		"rm": roomID,
	}).One(&match)
	if err != nil {
		return
	}

	// verify that player was in match
	host := (match.HostID == player.ID)
	guest := (match.OpponentID == player.ID)
	if !host && !guest {
		err = util.NewError("Player attempting to affect a match which they don't belong to")
	}

	// look for cached match result
	matchResult, foundResult := GetMatchResultByMatchId(context, roomID)

	// invert outcome and scores for guest
	if guest {
		outcome = invertOutcome(outcome)

		temp := playerScore
		playerScore = opponentScore
		opponentScore = temp
	}

	// check if opponent's result has already been submitted
	if foundResult {
		// validate outcome
		if outcome == MatchSurrender {
			// do nothing
		} else {
			log.Printf("Match result reconciliation: %v:%v %d:%d %d:%d", matchResult.Outcome, outcome, matchResult.HostScore, playerScore, matchResult.OpponentScore, opponentScore)
			
			if matchResult.Outcome == MatchSurrender || (matchResult.Outcome == outcome && matchResult.HostScore == playerScore && matchResult.OpponentScore == opponentScore) {
				match.State = MatchComplete
				match.Outcome = outcome
				match.HostScore = playerScore
				match.OpponentScore = opponentScore
			} else {
				match.State = MatchInvalid

				err = util.NewError("Non-symmetrical match outcomes reported by clients!")

				// TODO - remove victory tome from other player
			}

			// update match in database
			saveErr := match.Save(database)
			if saveErr != nil {
				log.Error(saveErr)
			}
		}

		ClearMatchResult(context, roomID)

		// after results are validated, process player stats for both players
		if match.State != MatchInvalid {
			err = match.ProcessMatchResults(database)
		}
	} else {
		// update results
		matchResult = &MatchResult {
			MatchID: match.ID,
			Outcome: outcome,
			HostScore: playerScore,
			OpponentScore: opponentScore,
		}

		// set results to cache
		SetMatchResult(context, roomID, matchResult)
	}

/*
	// prepare match change
	change := mgo.Change {
		Upsert: false,
		ReturnNew: true,
	}

	// check if host or guest
	var idField string
	if hosting {
		// prepare host query
		idField = "id1"

		// prepare host match change
		change.Update = bson.M {
			"$set": bson.M {
				"st": MatchCompleting,
				"oc": outcome,
				"s1": playerScore,
				"s2": opponentScore,
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
				"s1": playerScore,
				"s2": opponentScore,
				"t1": time.Now(),
			},
		}
	}

	// find active match for player, and update if found
	foundActiveMatch := true
	_, err = database.C(MatchCollectionName).Find(bson.M {
		idField: player.ID,
		"rm": roomID,
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
			"rm": roomID,
			"st": MatchCompleting,
		}).Apply(change, &match)
		if err != nil {
			if err.Error() == "not found" {
				err = util.NewError("Match not found")
			} else {
				err = util.NewError(err)
			}
			return
		}

		if match.Outcome == MatchSurrender || outcome == MatchSurrender {
			// surrender
		} else {
			// validate match outcome
			log.Printf("Match result reconciliation: %v:%v %d:%d %d:%d", match.Outcome, outcome, match.HostScore, playerScore, match.OpponentScore, opponentScore)
			if match.Outcome == outcome && match.HostScore == playerScore && match.OpponentScore == opponentScore {
				// match.State = MatchComplete
			} else {
				match.State = MatchInvalid

				// update as invalid
				match.Save(database)

				err = util.NewError("Non-symmetrical match outcomes reported by clients!")

				// TODO - roll back player stats!
			}
		}
	}
*/
	if match.State != MatchInvalid && err == nil && outcome != MatchSurrender {
		matchReward = &MatchReward {}

		if host {
			// player.ModifyArenaPoints(match.HostScore)
			matchReward.ArenaPoints = playerScore
		} else {
			// player.ModifyArenaPoints(match.OpponentScore)
			matchReward.ArenaPoints = opponentScore
		}

		if (host && outcome == MatchWin) || (!host && outcome == MatchLoss) {
			matchReward.Tome, err = player.AddVictoryTome(database)
		} else {
			//err = player.Save(database)
		}
	} 

	return
}

func (match *Match) GetHost(database *mgo.Database) (player *Player, err error) {
	if match.host == nil {
		if match.HostID.Valid() {
			match.host, err = GetPlayerById(database, match.HostID)
		}
	}
	return match.host, err
}

func (match *Match) GetOpponent(database *mgo.Database) (player *Player, err error) {
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
	host, err := match.GetHost(database)
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

		if rankChange > 0 || host.GetRankTier() > 1 {
			host.RankPoints += rankChange
		}
		if rankChange < 0 || opponent.GetRankTier() > 1 {
			opponent.RankPoints -= rankChange
		}

	case MatchElite:
		// get k-factor
		kFactor := getKFactor(host.Rating, opponent.Rating)

		// transformed ratings
		q1 := math.Pow10(host.Rating / 400)
		q2 := math.Pow10(opponent.Rating / 400)
		qs := q1 + q2

		// expected scores
		e1 := q1 / qs
		e2 := q2 / qs

		// observed scores
		s1 := 0.5 + float64(match.Outcome) * 0.5
		s2 := 1 - s1

		// resulting ratings
		r1 := host.Rating + util.RoundToInt(kFactor * (s1 - e1))
		r2 := opponent.Rating + util.RoundToInt(kFactor * (s2 - e2))

		//log.Printf("Elite Match Results: [%v(%v) + %v:%v => %v] vs. [%v(%v) + %v:%v => %v]", host.Rating, q1, e1, s1, r1, opponent.Rating, q2, e2, s2, r2)
		
		// update stats
		host.Rating = r1
		opponent.Rating = r2

	case MatchTournament:
		// TODO

	}

	// modify player stats
	host.MatchCount += 1
	opponent.MatchCount += 1
	switch match.Outcome {
	case MatchWin:
		host.WinCount += 1
		opponent.LossCount += 1
	case MatchLoss:
		host.LossCount += 1
		opponent.WinCount += 1
	}
	host.ModifyArenaPoints(match.HostScore)
	opponent.ModifyArenaPoints(match.OpponentScore)

	// update database
	err = host.Save(database)
	if err != nil {
		return
	}
	err = opponent.Save(database)
	return
}
