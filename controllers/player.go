package controllers

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
)

func HandlePlayer(application *system.Application) {
	application.HandleAPI("/player/set", system.TokenAuthentication, SetPlayer)
	//application.HandleAPI("/player/get", system.TokenAuthentication, GetPlayer)
}

func SetPlayer(context *system.Context) {
	// parse parameters
	data := context.Params.GetRequiredString("data")

	// update data
	player, err := models.UpdatePlayer(context.DB, context.User, data)
	if err != nil {
		panic(err)
	}

	// refresh cached name
	context.RefreshPlayerName(player)

	context.Message("Player updated successfully")
}

func GetPlayer(context *system.Context) {
	// get player
	player := context.GetPlayer()
	if player != nil {

		err := player.UpdateRewards(context.DB)
		if(err != nil) {
			panic(err)
		}
		
		// set successful response
		context.Message("Found player")
		context.Data = player
	} else {
		context.Fail(fmt.Sprintf("Failed to find player for username: %v", context.User.Username))
	}
}
