package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleMatch() {
	handleGameAPI("/match/clear", system.TokenAuthentication, MatchClear)
	handleGameAPI("/match/find", system.TokenAuthentication, MatchFind)
	handleGameAPI("/match/fail", system.TokenAuthentication, MatchFail)
	handleGameAPI("/match/result", system.TokenAuthentication, MatchResult)
}

func MatchClear(context *util.Context) {
	player := GetPlayer(context)

	// clear invalid matches
	util.Must(models.ClearMatches(context.DB, player))
}

func MatchFind(context *util.Context) {
	// parse parameters
	matchType := models.MatchType(context.Params.GetInt("type", int(models.MatchRanked)))

	player := GetPlayer(context)

	// find or queue match
	match, err := models.FindMatch(context.DB, player, matchType)
	util.Must(err)

	// respond
	context.SetData("match", match)
}

func MatchFail(context *util.Context) {
	player := GetPlayer(context)

	// fail any current match
	util.Must(models.FailMatch(context.DB, player))
}

func MatchResult(context *util.Context) {
	// parse parameters
	outcome := models.MatchOutcome(context.Params.GetRequiredInt("outcome"))
	playerScore := context.Params.GetRequiredInt("playerScore")
	opponentScore := context.Params.GetRequiredInt("opponentScore")
	roomID := context.Params.GetRequiredString("roomId")

	player := GetPlayer(context)
	
	// update match as complete
	_, reward, err := models.CompleteMatch(context, player, roomID, outcome, playerScore, opponentScore)
	util.Must(err)

	// get opponent
	// opponent, err := match.GetOpponent(context.DB)
	// util.Must(err)

	// update leaderboards
	updatePlayerPlace(context, player)
	// updatePlayerPlace(context, opponent)

	if reward != nil {
		player.SetDirty(models.PlayerDataMask_Tomes, models.PlayerDataMask_Stars)
		context.SetData("reward", reward)
	}
}
