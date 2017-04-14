package admin

import (
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
)

func handleAdminUsers(application *system.Application) {
	handleAdminTemplate(application, "/admin/users", system.TokenAuthentication, ShowUsers, "users.tmpl.html")
	handleAdminTemplate(application, "/admin/users/edit", system.TokenAuthentication, EditUser, "user.tmpl.html")
}

func ShowUsers(context *system.Context) {
	// parse parameters
	page := context.GetIntParameter("page", 1)

	// paginate users query
	query, pages, err := models.Paginate(context.Application.DB.C(models.UserCollectionName).Find(nil), DefaultPageSize, page)
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
	context.Set("page", page)
	context.Set("pages", pages)
}

func EditUser(context *system.Context) {
	// parse parameters
	userId := bson.ObjectIdHex(context.GetRequiredParameter("userId"))

	user, err := models.GetUserById(context.Application.DB, userId)
	if err != nil {
		panic(err)
	}

	player, err := models.GetPlayerByUser(context.Application.DB, userId)
	if err != nil {
		panic(err)
	}

	// handle request method
	switch context.Request.Method {
	case "POST":
		email := context.GetParameter("email", "")
		if email != "" {
			user.Email = email
			user.Update(context.Application.DB)
		}

		name := context.GetParameter("name", "")
		if name != "" {
			player.Name = name
		}

		standardCurrency := context.GetIntParameter("standardCurrency", -1)
		if standardCurrency >= 0 {
			player.StandardCurrency = standardCurrency
		}

		premiumCurrency := context.GetIntParameter("premiumCurrency", -1)
		if premiumCurrency >= 0 {
			player.PremiumCurrency = premiumCurrency
		}

		level := context.GetIntParameter("level", -1)
		if level >= 0 {
			player.Level = level
		}

		rating := context.GetIntParameter("rating", -1)
		if rating >= 0 {
			player.Rating = rating
		}

		stars := context.GetIntParameter("stars", -1)
		if stars >= 0 {
			player.Stars = stars
		}

		player.Update(context.Application.DB)
	}
	
	// set template bindings
	context.Data = user
	context.Set("user", user)
	context.Set("player", player)
}
