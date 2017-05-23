package admin

import (
	"fmt"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminUsers() {
	handleAdminTemplate("/admin/users", system.TokenAuthentication, ShowUsers, "users.tmpl.html")
	handleAdminTemplate("/admin/users/edit", system.TokenAuthentication, EditUser, "user.tmpl.html")
	handleAdminTemplate("/admin/users/reset", system.TokenAuthentication, ResetUser, "")
	handleAdminTemplate("/admin/users/delete", system.TokenAuthentication, DeleteUser, "")
}

func ShowUsers(context *util.Context) {
	// parse parameters
	search := context.Params.GetString("search", "")

	// process search terms
	var query *mgo.Query = nil
	if search != "" {
		// build query
		query = context.DB.C(models.UserCollectionName).Find(bson.M {
			"nm": bson.M {
				"$regex": bson.RegEx {
					Pattern: fmt.Sprintf(".*%s.*", search),
					Options: "i",
				},
			},
		})
	} else {
		query = context.DB.C(models.UserCollectionName).Find(nil)
	}

	// sorting
	query = context.Sort(query)

	// paginate users query
	pagination, err := context.Paginate(query, DefaultPageSize)
	util.Must(err)

	// get resulting users
	var users []*models.User
	util.Must(pagination.All(&users))

	// set template bindings
	context.Params.Set("users", users)
}

func EditUser(context *util.Context) {
	// parse parameters
	userId := context.Params.GetRequiredId("userId")

	user, err := models.GetUserById(context.DB, userId)
	util.Must(err)

	player, err := models.GetPlayerByUser(context.DB, userId)
	if err != nil {
		if err.Error() != "not found" {
			panic(err)
		}
	}

	// handle request method
	switch context.Request.Method {
	case "POST":
		userUpdated := false

		tag := context.Params.GetString("tag", "")
		if tag != "" {
			user.Tag = tag
			userUpdated = true
		}

		name := context.Params.GetString("name", "")
		if tag != "" {
			user.Name = name
			userUpdated = true
		}

		if userUpdated {
			user.Save(context.DB)
		}

		if player != nil {
			standardCurrency := context.Params.GetInt("standardCurrency", -1)
			if standardCurrency >= 0 {
				player.StandardCurrency = standardCurrency
			}

			premiumCurrency := context.Params.GetInt("premiumCurrency", -1)
			if premiumCurrency >= 0 {
				player.PremiumCurrency = premiumCurrency
			}

			xp := context.Params.GetInt("xp", -1)
			if xp >= 0 {
				player.XP = xp
			}

			rating := context.Params.GetInt("rating", -1)
			if rating >= 0 {
				player.Rating = rating
			}

			rankPoints := context.Params.GetInt("rankPoints", -1)
			if rankPoints >= 0 {
				player.RankPoints = rankPoints
			}

			arenaPoints := context.Params.GetInt("arenaPoints", -1)
			if arenaPoints >= 0 {
				player.ArenaPoints = arenaPoints
			}

			winCount := context.Params.GetInt("winCount", -1)
			if winCount >= 0 {
				player.WinCount = winCount
			}

			lossCount := context.Params.GetInt("lossCount", -1)
			if lossCount >= 0 {
				player.LossCount = lossCount
			}

			matchCount := context.Params.GetInt("matchCount", -1)
			if matchCount >= 0 {
				player.MatchCount = matchCount
			}

			player.Save(context.DB)
		}

		context.Message("Player updated!")
	}
	
	// set template bindings
	context.Params.Set("user", user)
	context.Params.Set("player", player)
}

func ResetUser(context *util.Context) {
	// parse parameters
	userId := context.Params.GetId("userId")

	if userId.Valid() {
		player, err := models.GetPlayerByUser(context.DB, userId)
		util.Must(err)

		player.Reset(context.DB)

		context.Redirect(fmt.Sprintf("/admin/users/edit?userId=%s", userId.Hex()), 302)
	} else {
		// get all players
		var players []*models.Player
		context.DB.C(models.PlayerCollectionName).Find(nil).All(&players)

		for _, player := range players {
			player.Reset(context.DB)
		}

		context.Redirect("/admin/users", 302)
	}
}

func DeleteUser(context *util.Context) {
	// parse parameters
	userId := context.Params.GetRequiredId("userId")
	page := context.Params.GetInt("page", 1)

	user, err := models.GetUserById(context.DB, userId)
	util.Must(err)

	player, err := models.GetPlayerByUser(context.DB, userId)
	util.Must(err)

	user.Delete(context.DB)
	player.Delete(context.DB)

	context.Redirect(fmt.Sprintf("/admin/users?page=%d", page), 302)
}
