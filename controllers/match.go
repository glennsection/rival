package controllers

import (
	"math"
	//"time"
	//"log"

	//"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
)

func HandleMatch(application *system.Application) {
	application.HandleAPI("/match/result", system.TokenAuthentication, MatchResult)
}

func MatchResult(context *system.Context) {
	// parse parameters
	match := &models.Match {}
	context.GetRequiredJSONParameter("match", match)
	// HACK
	// match := &models.Match {
	// 	PlayerID: bson.ObjectIdHex("58e817a17314dc0004bad2ad"),
	// 	OpponentID: bson.ObjectIdHex("58e7c02b3280475dc9a4f9ae"),
	// 	Outcome: models.MatchWin,
	// 	Time: time.Now(),
	// }

	// TODO - Validate that match.GetPlayer().UserID == context.User.ID

	// TODO - Attempt to match with opponent's call to this API.
	// If no match found in DB, then store this match and wait.
	// If no match arrives after X seconds or if outcomes don't sync, we have an error.
	// If match found in time, then call ProcessMatchResults(...).
	// HACK - for now just process winners
	if (match.Outcome == models.MatchWin) {
		ProcessMatchResults(context, match)
	}

	context.Message("Thanks for playing!")
}

// TODO - move to utils package
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TODO - move to utils package
func round(f float64) int {
    if f < -0.5 {
        return int(f - 0.5)
    }
    if f > 0.5 {
        return int(f + 0.5)
    }
    return 0
}

func GetKFactor(playerRating int, opponentRating int) float64 {
	// chess k-factors  (TODO - work on this...)
	rating := min(playerRating, opponentRating)
	if rating < 2100 {
		return 32.0
	} else if rating < 2400 {
		return 24.0
	}
	return 16.0
}

func ProcessMatchResults(context *system.Context, match *models.Match) {
	// get players
	player, err := match.GetPlayer(context.Application.DB)
	if err != nil {
		panic(err)
	}
	opponent, err := match.GetOpponent(context.Application.DB)
	if err != nil {
		panic(err)
	}

	// get k-factor
	kFactor := GetKFactor(player.Rating, opponent.Rating)

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
	r1 := player.Rating + int(round(kFactor * (s1 - e1)))
	r2 := opponent.Rating + int(round(kFactor * (s2 - e2)))

	//log.Printf("Match Results: [%v(%v) + %v:%v => %v] vs. [%v(%v) + %v:%v => %v]", player.Rating, q1, e1, s1, r1, opponent.Rating, q2, e2, s2, r2)
	
	// update stats
	player.Rating = r1
	opponent.Rating = r2
	player.MatchCount += 1
	opponent.MatchCount += 1
	switch match.Outcome {
	case models.MatchWin:
		player.WinCount += 1
		opponent.LossCount += 1
	case models.MatchLoss:
		player.LossCount += 1
		opponent.WinCount += 1
	}

	// update database
	player.Update(context.Application.DB)
	opponent.Update(context.Application.DB)
}