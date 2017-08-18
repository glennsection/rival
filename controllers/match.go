package controllers

import (
	"gopkg.in/mgo.v2/bson"

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
	util.Must(models.ClearMatches(context, []bson.ObjectId { player.ID }, models.MatchOpen))
}

func MatchFind(context *util.Context) {
	// parse parameters
	matchTypeName := context.Params.GetString("type", "Ranked")

	matchType := models.GetMatchType(matchTypeName)

	player := GetPlayer(context)

	// find or queue match
	match, err := models.FindPublicMatch(context, player.ID, matchType)
	util.Must(err)

	// respond
	context.SetData("match", match)
}

func MatchFail(context *util.Context) {
	player := GetPlayer(context)

	// fail any current match
	util.Must(models.FailMatch(context, player.ID))
}

func MatchResult(context *util.Context) {
	// parse parameters
	outcome := models.MatchOutcome(context.Params.GetRequiredInt("outcome"))
	playerScore := context.Params.GetRequiredInt("playerScore")
	opponentScore := context.Params.GetRequiredInt("opponentScore")
	roomID := context.Params.GetRequiredString("roomId")

	player := GetPlayer(context)

	// remember previous rank
	context.SetData("previousRankPoints", player.RankPoints)
	
	// update match as complete
	_, reward, err := models.CompleteMatch(context, player, roomID, outcome, playerScore, opponentScore)
	util.Must(err)

	if reward != nil {
		player.SetDirty(models.PlayerDataMask_Tomes, models.PlayerDataMask_Stars, models.PlayerDataMask_Quests, models.PlayerDataMask_Cards)
		context.SetData("reward", reward)
	}
}
