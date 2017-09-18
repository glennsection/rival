package models

import (
	"fmt"
	"time"
	"math"
	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/config"
	"bloodtales/data"
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
	MatchPrivate
	MatchActive
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
	GuestID         bson.ObjectId `bson:"id2,omitempty" json:"-"`
	Type            MatchType     `bson:"tp" json:"-"`
	RoomID          string        `bson:"rm" json:"roomId"`
	Arena           string        `bson:"ar" json:"arena"`
	State           MatchState    `bson:"st" json:"state"`
	Outcome       	MatchOutcome  `bson:"oc" json:"outcome"`
	HostScore       int           `bson:"s1" json:"hostScore"`
	GuestScore      int           `bson:"s2" json:"guestScore"`
	StartTime       time.Time     `bson:"t0" json:"-"`
	EndTime	        time.Time     `bson:"t1" json:"-"`

	// client
	Hosting         bool          `bson:"-" json:"hosting"`

	// internal
	host            *Player
	guest           *Player
}

// client model
type MatchClientAlias Match
type MatchClient struct {
	State           string        `json:"state"`

	*MatchClientAlias
}

// matchmaking ticket
type MatchTicket struct {
	PlayerID        string        `json:"pid"`
	MMR             int           `json:"mmr"`
	Type            MatchType     `json:"tp"`
	StartTime       time.Time     `json:"t0"`
	MatchID         string        `json:"mid"`
}

// cached match player results
type MatchPlayerResult struct {
	Score           int           `json:"s"`
	RankPoints      int           `json:"p"`
	Rating          int           `json:"r"`
}

// cached match results
type MatchResult struct {
	MatchID         bson.ObjectId     `json:"mid"`
	Outcome         MatchOutcome      `json:"oc"`
	Host            MatchPlayerResult `json:"p1"`
	Guest           MatchPlayerResult `json:"p2"`
}

// client rewards
type MatchReward struct {
	Tome            *Tome         `json:"tome"`
	TomeIndex       int           `json:"tomeIndex"`
	ArenaPoints     int           `json:"arenaPoints"`
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

func GetMatchType(matchName string) MatchType {
	switch matchName {
	case "Unranked":
		return MatchUnranked
	case "Ranked":
		return MatchRanked
	case "Elite":
		return MatchElite
	case "Tournament":
		return MatchTournament
	default:
		return MatchRanked
	}
}

func GetMatchTypeName(matchType MatchType) string {
	switch matchType {
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

func GetMatchById(context *util.Context, id bson.ObjectId) (match *Match, err error) {
	err = context.DB.C(MatchCollectionName).Find(bson.M { "_id": id } ).One(&match)
	return
}

func (match *Match) Save(context *util.Context) (err error) {
	if !match.ID.Valid() {
		match.ID = bson.NewObjectId()
	}

	// update match in database
	_, err = context.DB.C(MatchCollectionName).Upsert(bson.M { "_id": match.ID }, match)
	return
}

func (match *Match) Delete(context *util.Context) (err error) {
	return context.DB.C(MatchCollectionName).Remove(bson.M { "_id": match.ID })
}

func (matchResult *MatchResult) String() string {
	raw, err := json.Marshal(matchResult)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(raw)
}

func GetMatchResult(context *util.Context, roomID string) (matchResult *MatchResult, ok bool) {
	// get cache key
	key := fmt.Sprintf("MatchResult:%s", roomID)

	// get cached result
	ok = context.Cache.GetJSON(key, &matchResult)
	return
}

func SetMatchResult(context *util.Context, roomID string, matchResult *MatchResult) {
	// get cache key
	key := fmt.Sprintf("MatchResult:%s", roomID)

	// set cached result
	context.Cache.Set(key, matchResult)

	// expire temp results after some time
	if matchResult != nil {
		context.Cache.Expire(key, config.Config.Matches.MatchResultExpire)
	}
}

func ClearMatchResult(context *util.Context, roomID string) {
	SetMatchResult(context, roomID, nil)
}

func (ticket *MatchTicket) String() string {
	raw, err := json.Marshal(ticket)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(raw)
}

func GetMatchTicket(context *util.Context, playerID bson.ObjectId) (ticket *MatchTicket, ok bool) {
	// get cache key
	key := fmt.Sprintf("MatchTicket:%s", playerID.Hex())

	// get cached ticket
	ok = context.Cache.GetJSON(key, &ticket)
	return
}

func AddMatchTicket(context *util.Context, ticket *MatchTicket) {
	// get cache key
	key := fmt.Sprintf("MatchTicket:%s", ticket.PlayerID)

	// set cached ticket
	context.Cache.Set(key, ticket)

	// expire temp ticket after some time
	context.Cache.Expire(key, config.Config.Matches.MatchTicketExpire)

	// set MMR score
	context.Cache.SetScore("MMR", ticket.PlayerID, ticket.MMR)
}

func ClearMatchTicket(context *util.Context, playerID bson.ObjectId) {
	// get cache key
	key := fmt.Sprintf("MatchTicket:%s", playerID.Hex())

	// clear cached ticket
	context.Cache.Set(key, nil)

	// remove MMR score
	context.Cache.RemoveScore("MMR", playerID.Hex())
}

func ClearMatches(context *util.Context, playerIDs []bson.ObjectId, states ...MatchState) (err error) {
	// find and remove all invalid matches with player
	_, err = context.DB.C(MatchCollectionName).RemoveAll(bson.M {
		"$or": []bson.M {
			bson.M { "id1": bson.M { "$in": playerIDs }, },
			bson.M { "id2": bson.M { "$in": playerIDs }, },
		},
		"st": bson.M {
			"$in": states,
		},
 	})
 	return
}

func StartPrivateMatch(context *util.Context, hostID bson.ObjectId, guestID bson.ObjectId, matchType MatchType, roomID string, arenaName string) (match *Match, err error) {
	// check for existing match (TODO - verify that no room exists with this ID)
	err = ClearMatches(context, []bson.ObjectId { hostID, guestID }, MatchOpen, MatchActive)
	if err != nil {
		return
	}

	//log.Printf("StartPrivateMatch(%v, %v, %v, %v) => %v", hostID, guestID, matchType, roomID, match)

	// queue new match
	match = &Match {
		HostID: hostID,
		GuestID: guestID,
		Type: matchType,
		RoomID: roomID,
		Arena: arenaName,
		State: MatchActive,
		StartTime: time.Now(),
	}

	// clear any match tickets
	ClearMatchTicket(context, hostID)
	ClearMatchTicket(context, guestID)

	// update database
	err = match.Save(context)
	return
}

func FindPublicMatch(context *util.Context, player *Player, matchType MatchType) (match *Match, err error) {
	// check for existing match ticket for player
	ticket, _ := GetMatchTicket(context, player.ID)

	// check if we need to (re)create this ticket
	if ticket == nil || ticket.Type != matchType {
		// verify that no other pending matches exist for player
		err = ClearMatches(context, []bson.ObjectId { player.ID }, MatchOpen, MatchActive)
		if err != nil {
			return
		}

		// register match ticket for player
		ticket = &MatchTicket {
			PlayerID: player.ID.Hex(),
			MMR: player.GetMMR(),
			Type: matchType,
			StartTime: time.Now(),
			MatchID: "",
		}
	} else {
		// check if this ticket has a match reserved
		if ticket.MatchID != "" {
			// get match from database
			match, err = GetMatchById(context, bson.ObjectIdHex(ticket.MatchID))
			if err != nil {
				return
			}

			// make sure match still exists
			if match != nil {
				// clear current ticket
				ClearMatchTicket(context, player.ID)
				return
			} else {
				// clear match reservation in ticket
				ticket.MatchID = ""
			}
		}
	}
	// add or refresh ticket
	AddMatchTicket(context, ticket)

	// calculate desired max MMR delta based on time elapsed searching for match
	durationSearching := time.Now().Sub(ticket.StartTime)
	secondsSearching := int(durationSearching.Seconds())
	maxMMRDeltas := config.Config.Matches.MaxMMRDeltas
	maxMMRDelta := 0
	for i := 0; i < len(maxMMRDeltas); i++ {
		deltaTime := maxMMRDeltas[i]
		i++
		deltaMax := maxMMRDeltas[i]
		if secondsSearching < deltaTime {
			break
		}
		maxMMRDelta = deltaMax
	}

	// find matchmaking players within MMR window
	window := 1 // NOTE FIXME? - should only have to get the players directly adjacent to requesting,
	// since those are the closest MMRs.  Unless we have other criteria for matching beyond MMR...
	playerPlace := context.Cache.GetRank("MMR", player.ID.Hex())
	opponentIds := context.Cache.GetRankRange("MMR", playerPlace - window, playerPlace + window)

	// iterate through window to find optimal MMR opponent
	optimalOpponentId := ""
	optimalMMRDelta := maxMMRDelta + 1
	for _, opponentId := range opponentIds {
		if opponentId != player.ID.Hex() {
			opponentMMR := context.Cache.GetScore("MMR", opponentId)
			mmrDelta := (opponentMMR - ticket.MMR)
			if mmrDelta < 0 {
				mmrDelta = -mmrDelta
			}
			if mmrDelta < optimalMMRDelta {
				optimalMMRDelta = mmrDelta
				optimalOpponentId = opponentId
			}
		}
	}

	// check if opponent was found
	if optimalOpponentId != "" {
		matchOpponentId := bson.ObjectIdHex(optimalOpponentId)

		// get opponent ticket
		if opponentTicket, ok := GetMatchTicket(context, matchOpponentId); ok {
			// create match
			match = &Match {
				HostID: player.ID,
				GuestID: matchOpponentId,
				Type: matchType,
				RoomID: util.GenerateUUID(),
				Arena: data.GetRandomArena(),
				State: MatchActive,
				StartTime: time.Now(),
				Hosting: true,
			}

			// update match in database
			err = match.Save(context)
			if err != nil {
				return
			}

			// clear current player ticket
			ClearMatchTicket(context, player.ID)

			// update opponent ticket reservations for match (TODO - could use sockets here too...)
			opponentTicket.MatchID = match.ID.Hex()
			AddMatchTicket(context, opponentTicket)
		} else {
			log.Errorf("Failed to find matchmaking ticket for opponent player: %v", optimalOpponentId)
		}
	}
	return
}

func FailMatch(context *util.Context, playerID bson.ObjectId) (err error) {
	// find and remove all invalid matches with player
	var matches []*Match
	err = context.DB.C(MatchCollectionName).Find(bson.M {
		"$or": []bson.M {
			bson.M { "id1": playerID, },
			bson.M { "id2": playerID, },
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
				if match.HostID == playerID {
					match.HostID = match.GuestID
				}
				match.GuestID = bson.ObjectId("")
				match.State = MatchOpen
				match.Save(context)

				// TODO FIXME - need to send (via websocket), to new host, the fact that they are now the host
			} else {
				match.Delete(context)
			}
		}
	}
	return
}

func CompleteMatch(context *util.Context, player *Player, roomID string, outcome MatchOutcome, playerScore int, opponentScore int) (match *Match, matchReward *MatchReward, err error) {
	// get match from database
	err = context.DB.C(MatchCollectionName).Find(bson.M {
		"rm": roomID,
	}).One(&match)
	if err != nil {
		return
	}

	// verify that player was in match
	isHost := (match.HostID == player.ID)
	isGuest := (match.GuestID == player.ID)
	if !isHost && !isGuest {
		err = util.NewError("Player attempting to affect a match which they don't belong to")
		return
	}
	isWinner := (outcome == MatchWin)
	isLoser := (outcome == MatchLoss)

	// look for cached match result
	matchResult, foundResult := GetMatchResult(context, roomID)

	// assign scores
	hostScore := playerScore
	guestScore := opponentScore

	// invert outcome and scores for guest
	if isGuest {
		outcome = invertOutcome(outcome)

		hostScore = opponentScore
		guestScore = playerScore
	}

	// check if opponent's result has already been submitted
	if foundResult {
		// complete match
		match.State = MatchComplete
		match.EndTime = time.Now()

		// validate outcome
		if outcome == MatchSurrender {
			// use opponent's results
			match.Outcome = matchResult.Outcome
			match.HostScore = matchResult.Host.Score
			match.GuestScore = matchResult.Guest.Score
		} else {
			//log.Printf("Match result reconciliation: %v:%v %d:%d %d:%d", matchResult.Outcome, outcome, matchResult.Host.Score, hostScore, matchResult.Guest.Score, guestScore)
			
			if matchResult.Outcome == MatchSurrender || (matchResult.Outcome == outcome && matchResult.Host.Score == hostScore && matchResult.Guest.Score == guestScore) {
				// store results in match
				match.Outcome = outcome
				match.HostScore = hostScore
				match.GuestScore = guestScore
			} else {
				// invalid match
				match.State = MatchInvalid

				err = util.NewError("Non-symmetrical match outcomes reported by clients!")

				// TODO - revert stats for other player
			}
		}

		// update match in database
		saveErr := match.Save(context)
		if saveErr != nil {
			log.Error(saveErr)
		}

		// clear results in cache
		ClearMatchResult(context, roomID)
	} else {
		// get players
		var host *Player
		host, err = match.GetHost(context)
		if err != nil {
			return
		}
		var guest *Player
		guest, err = match.GetGuest(context)
		if err != nil {
			return
		}

		// process results
		matchResult = match.ProcessMatchResults(outcome, host, guest, hostScore, guestScore)

		// set results to cache
		SetMatchResult(context, roomID, matchResult)
	}

	// check that all is well, and update player in database
	if match.State != MatchInvalid && err == nil {
		matchReward = &MatchReward {}
		var playerResults *MatchPlayerResult

		// modify player stats and add tome
		player.MatchCount += 1
		if isHost {
			playerResults = &matchResult.Host
		} else {
			playerResults = &matchResult.Guest
		}

		var opponent *Player
		if isHost {
			opponent, err = match.GetGuest(context)
		} else {
			opponent, err = match.GetHost(context)
		}

		if err == nil { // add cards in the opponent's deck to the uncollected card slice
			player.AddUncollectedCards(opponent.Decks[opponent.CurrentDeck])
		}

		if outcome != MatchSurrender {
			previousArenaPoints := player.ArenaPoints
			player.ModifyArenaPoints(playerResults.Score)
			matchReward.ArenaPoints = player.ArenaPoints - previousArenaPoints
		}

		if isWinner {
			matchReward.TomeIndex, matchReward.Tome = player.AddVictoryTome(context)
			player.WinCount += 1
			player.UpdateDeckVictoryStats()
		} else if (isLoser) {
			player.LossCount += 1
		}
		player.RankPoints += playerResults.RankPoints
		player.Rating += playerResults.Rating

		// update battle quests
		player.UpdateQuests(nil, data.QuestTypeBattle, data.QuestTypeSinglePlayerBattle)

		// save player
		err = player.Save(context)

		// update cached player leaderboard place
		player.UpdatePlace(context)
	} 
	return
}

func (match *Match) ProcessMatchResults(outcome MatchOutcome, host *Player, guest *Player, hostScore int, guestScore int) (matchResult *MatchResult) {
	matchResult = &MatchResult {
		MatchID: match.ID,
		Outcome: outcome,
	}
	matchResult.Host.Score = hostScore
	matchResult.Guest.Score = guestScore

	// update according to match type
	switch match.Type {

	case MatchRanked, MatchTournament: // TODO - tournaments should use different ranking eventually
		// update stats
		rankChange := int(outcome)

		if (rankChange == -2) { // handle surrender - players should never lose more than one rank
			rankChange = -1
		}

		if rankChange > 0 || host.GetRankTier() > 1 {
			matchResult.Host.RankPoints = rankChange
		}
		if rankChange < 0 || guest.GetRankTier() > 1 {
			matchResult.Guest.RankPoints = -rankChange
		}

	case MatchElite:
		// get k-factor
		kFactor := getKFactor(host.Rating, guest.Rating)

		// transformed ratings
		q1 := math.Pow10(host.Rating / 400)
		q2 := math.Pow10(guest.Rating / 400)
		qs := q1 + q2

		// expected scores
		e1 := q1 / qs
		e2 := q2 / qs

		// observed scores
		s1 := 0.5 + float64(match.Outcome) * 0.5
		s2 := 1 - s1

		// resulting ratings
		r1 := util.RoundToInt(kFactor * (s1 - e1))
		r2 := util.RoundToInt(kFactor * (s2 - e2))

		//log.Printf("Elite Match Results: [%v(%v) + %v:%v => %v] vs. [%v(%v) + %v:%v => %v]", host.Rating, q1, e1, s1, r1, guest.Rating, q2, e2, s2, r2)
		
		// update stats
		matchResult.Host.Rating = r1
		matchResult.Guest.Rating = r2

	//case MatchTournament:
		// TODO

	}
	return
}

func (match *Match) GetHost(context *util.Context) (player *Player, err error) {
	if match.host == nil {
		if match.HostID.Valid() {
			match.host, err = GetPlayerById(context, match.HostID)
		} else {
			err = util.NewError(fmt.Sprintf("Invalid Host set in match: %s", match.RoomID))
		}
	}
	return match.host, err
}

func (match *Match) GetGuest(context *util.Context) (player *Player, err error) {
	if match.guest == nil {
		if match.GuestID.Valid() {
			match.guest, err = GetPlayerById(context, match.GuestID)
		} else {
			err = util.NewError(fmt.Sprintf("Invalid Guest set in match: %s", match.RoomID))
		}
	}
	return match.guest, err
}

func (match *Match) GetTypeName() string {
	return GetMatchTypeName(match.Type)
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

func (match *Match) GetOutcomeName() string {
	switch match.State {
	default:
		return "-"
	case MatchComplete:
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
		return MatchWin
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
