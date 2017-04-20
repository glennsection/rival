package admin

import (
	"bloodtales/system"
	"bloodtales/models"
)

type Analytics struct {
	// TODO
}

func handleAdminAnalytics(application *system.Application) {
	handleAdminTemplate(application, "/admin/leaderboard", system.NoAuthentication, ShowLeaderboard, "leaderboard.tmpl.html")
}

func ShowLeaderboard(context *system.Context) {
	// parse parameters
	page := context.Params.GetInt("page", 1)

	// paginate players query (TODO - use redis!)
	pagination, err := system.Paginate(context.DB.C(models.PlayerCollectionName).Find(nil).Sort("-rk"), DefaultPageSize, page)
	if err != nil {
		panic(err)
	}

	// get resulting players
	var players []models.Player
	err = pagination.All(&players)
	if err != nil {
		panic(err)
	}

	// set template bindings
	context.Data = players
	context.Params.Set("pagination", pagination)
}
