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
	player := context.GetPlayer()

	// clear invalid matches
	err := models.ClearMatches(context.DB, player)
	if err != nil {
		panic(err)
	}
}

func MatchFind(context *system.Context) {
	// parse parameters
	matchType := models.MatchType(context.Params.GetInt("type", int(models.MatchRanked)))

	player := context.GetPlayer()

	// find or queue match
	match, err := models.FindMatch(context.DB, player, matchType)
	if err != nil {
		panic(err)
	}

	// respond
	context.Data = match
}

func MatchFail(context *system.Context) {
	player := context.GetPlayer()

	// fail any current match
	err := models.FailMatch(context.DB, player)
	if err != nil {
		panic(err)
	}
}

func MatchResult(context *system.Context) {
	// parse parameters
	outcome := models.MatchOutcome(context.Params.GetRequiredInt("outcome"))

	player := context.GetPlayer()
	
	// update match as complete
	reward, err := models.CompleteMatch(context.DB, player, outcome)
	if err != nil {
		panic(err)
	}

	if reward != nil {
		context.Message("Congratulations! You earned a tome!")
		context.Data = reward
	}

	context.Message("Thanks for playing!")
}
