package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
)

func HandleMatch(application *system.Application) {
	application.HandleAPI("/match/clear", system.TokenAuthentication, MatchClear)
	application.HandleAPI("/match/find", system.TokenAuthentication, MatchFind)
	application.HandleAPI("/match/fail", system.TokenAuthentication, MatchFail)
	application.HandleAPI("/match/result", system.TokenAuthentication, MatchResult)
}

func MatchClear(context *system.Context) {
	player := GetPlayer(context)

	// clear invalid matches
	err := models.ClearMatches(context.DB, player)
	if err != nil {
		panic(err)
	}
}

func MatchFind(context *system.Context) {
	// parse parameters
	matchType := models.MatchType(context.Params.GetInt("type", int(models.MatchRanked)))

	player := GetPlayer(context)

	// find or queue match
	match, err := models.FindMatch(context.DB, player, matchType)
	if err != nil {
		panic(err)
	}

	// respond
	context.Data = match
}

func MatchFail(context *system.Context) {
	player := GetPlayer(context)

	// fail any current match
	err := models.FailMatch(context.DB, player)
	if err != nil {
		panic(err)
	}
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
	if err != nil {
		panic(err)
	}

	// get opponent
	opponent, err := match.GetOpponent(context.DB)
	if err != nil {
		panic(err)
	}

	// update leaderboards
	updatePlayerPlace(context, player)
	updatePlayerPlace(context, opponent)

	if reward != nil {
		context.Data = reward
	}

	context.Message("Thanks for playing!")
}
