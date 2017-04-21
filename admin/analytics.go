package admin

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
)

func handleAdminAnalytics(application *system.Application) {
	handleAdminTemplate(application, "/admin/leaderboard", system.NoAuthentication, ShowLeaderboard, "leaderboard.tmpl.html")

	handleAdminTemplate(application, "/admin/matches", system.TokenAuthentication, ShowMatches, "matches.tmpl.html")
	handleAdminTemplate(application, "/admin/matches/edit", system.TokenAuthentication, EditMatch, "match.tmpl.html")
	handleAdminTemplate(application, "/admin/matches/delete", system.TokenAuthentication, DeleteMatch, "")
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
	for _, match := range matches {
		err = match.LoadPlayers(context.DB)
		if err != nil {
			panic(err)
		}
	}

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
	err = match.LoadPlayers(context.DB)
	if err != nil {
		panic(err)
	}

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