package admin

import (
	"fmt"
	
	"bloodtales/system"
	"bloodtales/models"
)

func handleAdminUsers(application *system.Application) {
	handleAdminTemplate(application, "/admin/users", system.TokenAuthentication, ShowUsers, "users.tmpl.html")
	handleAdminTemplate(application, "/admin/users/edit", system.TokenAuthentication, EditUser, "user.tmpl.html")
	handleAdminTemplate(application, "/admin/users/reset", system.TokenAuthentication, ResetUser, "")
	handleAdminTemplate(application, "/admin/users/delete", system.TokenAuthentication, DeleteUser, "")
}

func ShowUsers(context *system.Context) {
	// paginate users query
	pagination, err := context.Paginate(context.DB.C(models.UserCollectionName).Find(nil), DefaultPageSize)
	if err != nil {
		panic(err)
	}

	// get resulting users
	var users []*models.User
	err = pagination.All(&users)
	if err != nil {
		panic(err)
	}

	// set template bindings
	context.Data = users
}

func EditUser(context *system.Context) {
	// parse parameters
	userId := context.Params.GetRequiredID("userId")

	user, err := models.GetUserById(context.DB, userId)
	if err != nil {
		panic(err)
	}

	player, err := models.GetPlayerByUser(context.DB, userId)
	if err != nil {
		if err.Error() != "not found" {
			panic(err)
		}
	}

	// handle request method
	switch context.Request.Method {
	case "POST":
		email := context.Params.GetString("email", "")
		if email != "" {
			user.Email = email
			user.Update(context.DB)
		}

		if player != nil {
			name := context.Params.GetString("name", "")
			if name != "" {
				player.Name = name
			}

			standardCurrency := context.Params.GetInt("standardCurrency", -1)
			if standardCurrency >= 0 {
				player.StandardCurrency = standardCurrency
			}

			premiumCurrency := context.Params.GetInt("premiumCurrency", -1)
			if premiumCurrency >= 0 {
				player.PremiumCurrency = premiumCurrency
			}

			level := context.Params.GetInt("level", -1)
			if level >= 0 {
				player.Level = level
			}

			rating := context.Params.GetInt("rating", -1)
			if rating >= 0 {
				player.Rating = rating
			}

			rankPoints := context.Params.GetInt("rankPoints", -1)
			if rankPoints >= 0 {
				player.RankPoints = rankPoints
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

			player.Update(context.DB)
		}

		context.Message("Player updated!")
	}
	
	// set template bindings
	context.Data = user
	context.Params.Set("user", user)
	context.Params.Set("player", player)
}

func ResetUser(context *system.Context) {
	// parse parameters
	userId := context.Params.GetID("userId")

	if userId.Valid() {
		player, err := models.GetPlayerByUser(context.DB, userId)
		if err != nil {
			panic(err)
		}

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

func DeleteUser(context *system.Context) {
	// parse parameters
	userId := context.Params.GetRequiredID("userId")
	page := context.Params.GetInt("page", 1)

	user, err := models.GetUserById(context.DB, userId)
	if err != nil {
		panic(err)
	}

	player, err := models.GetPlayerByUser(context.DB, userId)
	if err != nil {
		panic(err)
	}

	user.Delete(context.DB)
	player.Delete(context.DB)

	context.Redirect(fmt.Sprintf("/admin/users?page=%d", page), 302)
}
