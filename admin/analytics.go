package admin

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminAnalytics(application *system.Application) {
	handleAdminTemplate(application, "/admin/leaderboard", system.NoAuthentication, ShowLeaderboard, "leaderboard.tmpl.html")

	handleAdminTemplate(application, "/admin/matches", system.TokenAuthentication, ShowMatches, "matches.tmpl.html")
	handleAdminTemplate(application, "/admin/matches/edit", system.TokenAuthentication, EditMatch, "match.tmpl.html")
	handleAdminTemplate(application, "/admin/matches/delete", system.TokenAuthentication, DeleteMatch, "")
	handleAdminTemplate(application, "/admin/matches/reset", system.TokenAuthentication, ResetMatches, "")
}

func ShowLeaderboard(context *system.Context) {
	// parse parameters
	page := context.Params.GetInt("page", 1)

	// TODO - correct pagination
	pageStart := DefaultPageSize * (page - 1)
	pageStop := DefaultPageSize * page - 1
	playerIds := context.Cache.GetRankRange("Leaderboard", pageStart, pageStop)

	// convert to ObjectIds
	playerObjectIds := make([]bson.ObjectId, len(playerIds))
	for i, id := range playerIds {
		playerObjectIds[i] = bson.ObjectIdHex(id)
	}

	// get players
	var unsortedPlayers []*models.Player
	util.Must(context.DB.C(models.PlayerCollectionName).Find(bson.M {
		"_id": bson.M { "$in": playerObjectIds, },
	}).All(&unsortedPlayers))
	
	// reorder
	players := make([]*models.Player, len(unsortedPlayers))
	for _, player := range unsortedPlayers {
		for j, playerId := range playerObjectIds {
			if playerId == player.ID {
				players[j] = player
				break
			}
		}
	}


	// get players (TODO - aggregation if addFields worked)
	// var players []*models.Player
	// err := context.DB.C(models.PlayerCollectionName).Pipe([]bson.M {
	// 	bson.M {
	// 		"$match": bson.M {
	// 			"_id": bson.M { "$in": playerObjectIds, },
	// 		},
	// 	},
	// 	bson.M {
	// 		"$addFields": bson.M {
	// 			"__order": bson.M { "$indexOfArray": []interface{} { playerObjectIds, "$name" }, },
	// 		},
	// 	},
	// 	bson.M {
	// 		"$sort": bson.M {
	// 			"__order": 1,
	// 		},
	// 	},
	// }).All(&players)
	// if err != nil {
	// 	panic(err)
	// }



	// paginate players query
	// pagination, err := context.Paginate(context.DB.C(models.PlayerCollectionName).Find(nil).Sort("-rk"), DefaultPageSize)
	// if err != nil {
	// 	panic(err)
	// }

	// // get resulting players
	// var players []*models.Player
	// err = pagination.All(&players)
	// if err != nil {
	// 	panic(err)
	// }

	// set template bindings
	context.Data = players
}

func ShowMatches(context *system.Context) {
	// paginate players query (TODO - use redis!)
	pagination, err := context.Paginate(context.DB.C(models.MatchCollectionName).Find(nil).Sort("-t0"), DefaultPageSize)
	util.Must(err)

	// get resulting matches
	var matches []*models.Match
	util.Must(pagination.All(&matches))

	// set template bindings
	context.Data = matches
}

func EditMatch(context *system.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredID("matchId")

	match, err := models.GetMatchById(context.DB, matchId)
	util.Must(err)

	// handle request method
	switch context.Request.Method {
	case "POST":
		context.Message("Match updated!")
	}
	
	// set template bindings
	context.Data = match
}

func DeleteMatch(context *system.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredID("matchId")
	page := context.Params.GetInt("page", 1)

	match, err := models.GetMatchById(context.DB, matchId)
	util.Must(err)

	match.Delete(context.DB)

	context.Redirect(fmt.Sprintf("/admin/matches?page=%d", page), 302)
}

func ResetMatches(context *system.Context) {
	/*
	// get all players
	var players []*models.Player
	err := context.DB.C(models.PlayerCollectionName).Find(nil).All(&players)
	if err != nil {
		panic(err)
	}

	// reset all players (TODO - use db.Run() bulk API)
	for _, player := range players {
		player.Initialize()
		player.Save(context.DB)
	}
	*/

	// bulk reset of match data
	var result bson.D
	util.Must(context.DB.Run(bson.D {
		bson.DocElem { "update",  models.PlayerCollectionName },
		bson.DocElem { "updates",  []bson.M {
			bson.M {
				"q": bson.M {},
				"u": bson.M {
					"rk": 0,
					"rt": 1200,
					"wc": 0,
					"lc": 0,
					"mc": 0,
					"ap": 0,
				},
				"multi": false,
				"upsert": false,
				"limit": 0,
			},
		} },
		bson.DocElem { "writeConcern", bson.M {
			"w": 1,
			"j": true,
			"wtimeout": 1000,
		} },
		bson.DocElem { "ordered", false },
	}, &result))

	context.Redirect("/admin/leaderboard", 302)
}