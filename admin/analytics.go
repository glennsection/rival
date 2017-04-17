package admin

import (
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

type Analytics struct {
	// TODO
}

func handleAdminAnalytics(application *system.Application) {
	handleAdminTemplate(application, "/admin/leaderboard", system.TokenAuthentication, ShowLeaderboard, "leaderboard.tmpl.html")
}

func ShowLeaderboard(context *system.Context) {
	// parse parameters
	page := context.GetIntParameter("page", 1)

	// paginate players query (TODO - use redis!)
	query, pages, err := util.Paginate(context.DB.C(models.PlayerCollectionName).Find(nil).Sort("-rt"), DefaultPageSize, page)
	if err != nil {
		panic(err)
	}

	// get resulting players
	var players []models.Player
	err = query.All(&players)
	if err != nil {
		panic(err)
	}

	// set template bindings
	context.Data = players
	context.Set("page", page)
	context.Set("pages", pages)
}
