package admin

import (
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminLeaderboards() {
	handleAdminTemplate("/leaderboard", system.NoAuthentication, ViewLeaderboard, "leaderboard.tmpl.html")
	handleAdminTemplate("/admin/leaderboard/refresh", system.TokenAuthentication, RefreshLeaderboard, "")
}

func ViewLeaderboard(context *util.Context) {
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
	var players []*models.Player
	for _, playerId := range playerObjectIds {
		for _, player := range unsortedPlayers {
			if playerId == player.ID {
				players = append(players, player)
				break
			}
		}
	}


	// paginate players query
	// pagination, err := context.Paginate(context.DB.C(models.PlayerCollectionName).Find(nil).Sort("lb"), DefaultPageSize)
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
	context.Params.Set("players", players)
}

func RefreshLeaderboard(context *util.Context) {
	// parse parameters
	playerId := context.Params.GetId("playerId")

	if playerId.Valid() {
		player, err := models.GetPlayerById(context, playerId)
		util.Must(err)

		player.UpdatePlace(context)

		context.Redirect("/users/edit?userId=" + player.UserID.Hex(), 302)
	} else {
		// HACK - inefficient
		models.UpdateAllPlayersPlace(context)

		context.Redirect("/leaderboard", 302)
	}
}