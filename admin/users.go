package admin

import (
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
)

func handleAdminUsers(application *system.Application) {
	handleAdminTemplate(application, "/admin/users", system.TokenAuthentication, ShowUsers, "users.tmpl.html")
	handleAdminTemplate(application, "/admin/users/edit", system.TokenAuthentication, ShowUser, "user.tmpl.html")
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

func ShowUser(context *system.Context) {
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

	// set template bindings
	context.Data = user
	context.Set("user", user)
	context.Set("player", player)
}
