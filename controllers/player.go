package controllers

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
)

func HandlePlayer(application *system.Application) {
	application.HandleAPI("/player/set", system.TokenAuthentication, SetPlayer)
	application.HandleAPI("/player/get", system.TokenAuthentication, GetPlayer)
}

func SetPlayer(context *system.Context) {
	// parse parameters
	data := context.GetRequiredParameter("data")

	// update data
	if err := models.UpdatePlayer(context.DB, context.User, data); err != nil {
		panic(err)
	}

	context.Message("Player updated successfully")
}

func GetPlayer(context *system.Context) {
	// get player
	player := context.GetPlayer()
	if player != nil {
		// set successful response
		context.Message("Found player")
		context.Data = player
	} else {
		panic(fmt.Sprintf("Failed to find player for username: %v", context.User.Username))
	}
}
