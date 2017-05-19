package admin

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminMatches() {
	handleAdminTemplate("/admin/matches", system.TokenAuthentication, ShowMatches, "matches.tmpl.html")
	handleAdminTemplate("/admin/matches/edit", system.TokenAuthentication, EditMatch, "match.tmpl.html")
	handleAdminTemplate("/admin/matches/delete", system.TokenAuthentication, DeleteMatch, "")
	handleAdminTemplate("/admin/matches/reset", system.TokenAuthentication, ResetMatches, "")
}

func ShowMatches(context *util.Context) {
	// paginate players query (TODO - use redis!)
	pagination, err := context.Paginate(context.DB.C(models.MatchCollectionName).Find(nil).Sort("-t0"), DefaultPageSize)
	util.Must(err)

	// get resulting matches
	var matches []*models.Match
	util.Must(pagination.All(&matches))

	// set template bindings
	context.Data = matches
}

func EditMatch(context *util.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredId("matchId")

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

func DeleteMatch(context *util.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredId("matchId")
	page := context.Params.GetInt("page", 1)

	match, err := models.GetMatchById(context.DB, matchId)
	util.Must(err)

	match.Delete(context.DB)

	context.Redirect(fmt.Sprintf("/admin/matches?page=%d", page), 302)
}

func ResetMatches(context *util.Context) {
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