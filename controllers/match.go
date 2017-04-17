package controllers

import (
	"time"
	//"log"

	//"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
)

func HandleMatch(application *system.Application) {
	application.HandleAPI("/match/find", system.TokenAuthentication, MatchFind)
	application.HandleAPI("/match/result", system.TokenAuthentication, MatchResult)
}

func MatchFind(context *system.Context) {
	// parse parameters
	//matchType := models.MatchType(context.GetIntParameter("type", int(models.MatchRanked)))

	//match := &models.Match {}
	// TODO.. find match or queue
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
	// }

	// TODO - Validate that match.GetPlayer().UserID == context.User.ID

	// HACK - set match type here for now
	match.Type = models.MatchRanked

	// set current time
	match.Time = time.Now()

	// TODO - Attempt to match with opponent's call to this API.
	// If no match found in DB, then store this match and wait.
	// If no match arrives after X seconds or if outcomes don't sync, we have an error.
	// If match found in time, then call ProcessMatchResults(...).

	// HACK - for now just process winners
	if (match.Outcome == models.MatchWin) {
		match.ProcessMatchResults(context.DB)
	}

	context.Message("Thanks for playing!")
}
