package controllers

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
)

func HandlePlayer(application *system.Application) {
	application.HandleAPI("/player/set", system.TokenAuthentication, SetPlayer)
	application.HandleAPI("/player/name", system.TokenAuthentication, SetPlayerName)
	//application.HandleAPI("/player/get", system.TokenAuthentication, GetPlayer)
}

func SetPlayerName(context *system.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get player
	player := context.GetPlayer()

	// set name and update
	player.Name = name
	err := player.Update(context.DB)
	if err != nil {
		panic(err)
	}

	// refresh cached name
	context.RefreshPlayerName(player)
}

func SetPlayer(context *system.Context) {
	// parse parameters
	data := context.Params.GetRequiredString("data")

	// update data
	_, err := models.UpdatePlayer(context.DB, context.User, data)
	if err != nil {
		panic(err)
	}

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
		context.SetDirty([]int64{	models.UpdateMask_Name, 
									models.UpdateMask_Currency, 
									models.UpdateMask_XP, 
									models.UpdateMask_Cards, 
									models.UpdateMask_Deck,
									models.UpdateMask_Loadout,
									models.UpdateMask_Tomes,
									models.UpdateMask_Stars,
    								models.UpdateMask_Quests})
	} else {
		context.Fail(fmt.Sprintf("Failed to find player for username: %v", context.User.Username))
	}
}
