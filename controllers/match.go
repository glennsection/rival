package controllers

import (
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
)

func HandleMatch(application *system.Application) {
	application.HandleAPI("/match/clear", system.TokenAuthentication, MatchClear)
	application.HandleAPI("/match/find", system.TokenAuthentication, MatchFind)
	application.HandleAPI("/match/result", system.TokenAuthentication, MatchResult)
}

func MatchClear(context *system.Context) {
	player := context.GetPlayer()

	// find all invalid matches
	context.DB.C(models.MatchCollectionName).RemoveAll(bson.M {
		"$or": []bson.M {
			bson.M { "id1": player.ID, },
			bson.M { "id2": player.ID, },
		},
		"st": bson.M {
			"$in": []interface{} {
				models.MatchInvalid,
				models.MatchOpen,
				models.MatchActive,
			},
		},
 	})
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
