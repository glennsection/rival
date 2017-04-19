package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
)

func HandleMatch(application *system.Application) {
	application.HandleAPI("/match/find", system.TokenAuthentication, MatchFind)
	application.HandleAPI("/match/result", system.TokenAuthentication, MatchResult)
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

func MatchResult(context *system.Context) {
	// parse parameters
	outcome := models.MatchOutcome(context.Params.GetRequiredInt("outcome"))

	player := context.GetPlayer()
	
	// update match as complete
	err := models.CompleteMatch(context.DB, player, outcome)
	if err != nil {
		panic(err)
	}

	context.Message("Thanks for playing!")
}
