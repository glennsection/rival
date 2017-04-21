package admin

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/log"
)

func handleAdminAnalytics(application *system.Application) {
	handleAdminTemplate(application, "/admin/leaderboard", system.NoAuthentication, ShowLeaderboard, "leaderboard.tmpl.html")

	handleAdminTemplate(application, "/admin/matches", system.TokenAuthentication, ShowMatches, "matches.tmpl.html")
	handleAdminTemplate(application, "/admin/matches/edit", system.TokenAuthentication, EditMatch, "match.tmpl.html")
	handleAdminTemplate(application, "/admin/matches/delete", system.TokenAuthentication, DeleteMatch, "")
	handleAdminTemplate(application, "/admin/matches/reset", system.TokenAuthentication, ResetMatches, "")
}

func ShowLeaderboard(context *system.Context) {
	// paginate players query (TODO - use redis!)
	pagination, err := context.Paginate(context.DB.C(models.PlayerCollectionName).Find(nil).Sort("-rk"), DefaultPageSize)
	if err != nil {
		panic(err)
	}

	// get resulting players
	var players []*models.Player
	err = pagination.All(&players)
	if err != nil {
		panic(err)
	}

	// set template bindings
	context.Data = players
}

func ShowMatches(context *system.Context) {
	// paginate players query (TODO - use redis!)
	pagination, err := context.Paginate(context.DB.C(models.MatchCollectionName).Find(nil).Sort("-t0"), DefaultPageSize)
	if err != nil {
		panic(err)
	}

	// get resulting matches
	var matches []*models.Match
	err = pagination.All(&matches)
	if err != nil {
		panic(err)
	}

	// load players
	// for _, match := range matches {
	// 	err = match.LoadPlayers(context.DB)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// set template bindings
	context.Data = matches
}

func EditMatch(context *system.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredID("matchId")

	match, err := models.GetMatchById(context.DB, matchId)
	if err != nil {
		panic(err)
	}

	// load players
	// err = match.LoadPlayers(context.DB)
	// if err != nil {
	// 	panic(err)
	// }

	// handle request method
	switch context.Request.Method {
	case "POST":
		// matchCount := context.Params.GetInt("matchCount", -1)
		// if matchCount >= 0 {
		// 	player.MatchCount = matchCount
		// }

		// match.Update(context.DB)

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
	if err != nil {
		panic(err)
	}

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
		player.Update(context.DB)
	}
	*/

	// bulk reset of match data
	var result bson.D
	err := context.DB.Run(bson.D {
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
	}, &result)
	log.Printf("RESULT: %v", result)
	if err != nil {
		panic(err)
	}

	context.Redirect("/admin/leaderboard", 302)
}