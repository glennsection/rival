package admin

import (
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleAdminUsers(application *system.Application) {
	handleAdminTemplate(application, "/admin/users", system.TokenAuthentication, ShowUsers, "users.tmpl.html")
	handleAdminTemplate(application, "/admin/users/edit", system.TokenAuthentication, EditUser, "user.tmpl.html")
	handleAdminTemplate(application, "/admin/users/delete", system.TokenAuthentication, DeleteUser, "")
}

func ShowUsers(context *system.Context) {
	// parse parameters
	page := context.Params.GetInt("page", 1)

	// paginate users query
	query, pages, err := util.Paginate(context.DB.C(models.UserCollectionName).Find(nil), DefaultPageSize, page)
	if err != nil {
		panic(err)
	}

	// get resulting users
	var users []models.User
	err = query.All(&users)
	if err != nil {
		panic(err)
	}

	// set template bindings
	context.Data = users
	context.Params.Set("page", page)
	context.Params.Set("pages", pages)
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
		panic(err)
	}

	// handle request method
	switch context.Request.Method {
	case "POST":
		email := context.Params.GetString("email", "")
		if email != "" {
			user.Email = email
			user.Update(context.DB)
		}

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

		rank := context.Params.GetInt("rank", -1)
		if rank >= 0 {
			player.Rank = rank
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

		context.Message("Player updated!")
	}
	
	// set template bindings
	context.Data = user
	context.Params.Set("user", user)
	context.Params.Set("player", player)
}

func DeleteUser(context *system.Context) {
	// parse parameters
	userId := context.Params.GetRequiredID("userId")

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

	context.Redirect("/admin/users", 301)
}