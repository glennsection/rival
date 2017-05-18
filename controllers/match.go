package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func HandleMatch() {
	HandleGameAPI("/match/clear", system.TokenAuthentication, MatchClear)
	HandleGameAPI("/match/find", system.TokenAuthentication, MatchFind)
	HandleGameAPI("/match/fail", system.TokenAuthentication, MatchFail)
	HandleGameAPI("/match/result", system.TokenAuthentication, MatchResult)
}

func MatchClear(context *system.Context) {
	player := GetPlayer(context)

	// clear invalid matches
	util.Must(models.ClearMatches(context.DB, player))
}

func MatchFind(context *system.Context) {
	// parse parameters
	matchType := models.MatchType(context.Params.GetInt("type", int(models.MatchRanked)))

	player := GetPlayer(context)

	// find or queue match
	match, err := models.FindMatch(context.DB, player, matchType)
	util.Must(err)

	// respond
	context.Data = match
}

func MatchFail(context *system.Context) {
	player := GetPlayer(context)

	// fail any current match
	util.Must(models.FailMatch(context.DB, player))
}

func MatchResult(context *system.Context) {
	// parse parameters
	outcome := models.MatchOutcome(context.Params.GetRequiredInt("outcome"))
	playerScore := context.Params.GetRequiredInt("playerScore")
	opponentScore := context.Params.GetRequiredInt("opponentScore")
	host := context.Params.GetRequiredBool("host")

	player := GetPlayer(context)
	
	// update match as complete
	match, reward, err := models.CompleteMatch(context.DB, player, host, outcome, playerScore, opponentScore)
	util.Must(err)

	// get opponent
	opponent, err := match.GetOpponent(context.DB)
	util.Must(err)

	// update leaderboards
	updatePlayerPlace(context, player)
	updatePlayerPlace(context, opponent)

	if reward != nil {
		context.SetDirty([]int64{models.UpdateMask_Tomes,
								 models.UpdateMask_Stars})
		context.Data = reward
	}

	context.Message("Thanks for playing!")
}
