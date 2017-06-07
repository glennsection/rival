package controllers

import (
	"fmt"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handlePlayer() {
	handleGameAPI("/player/set", system.TokenAuthentication, OverwritePlayer) // HACK
	handleGameAPI("/player/name", system.TokenAuthentication, SetPlayerName)

	// template functions
	util.AddTemplateFunc("getUserName", models.GetUserName)
	util.AddTemplateFunc("getPlayerName", models.GetPlayerName)
}

func GetPlayer(context *util.Context) (player *models.Player) {
	// get player for current context, with cache in params
	player, ok := context.Params.Get("_player").(*models.Player)
	if ok == false {
		user := system.GetUser(context)
		if user != nil {
			player, _ = models.GetPlayerByUser(context, user.ID)

			if player != nil {
				context.Params.Set("_player", player)
			}
		}
	}
	return
}

func SetPlayerName(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get user
	user := system.GetUser(context)

	// set name and update
	user.Name = name
	err := user.Save(context)
	util.Must(err)

	// get player
	player, err := models.GetPlayerByUser(context, user.ID)
	util.Must(err)

	// update cache
	player.CacheName(context, name)
}

func OverwritePlayer(context *util.Context) {
	// parse parameters
	data := context.Params.GetRequiredString("data")

	// update data
	player := GetPlayer(context)
	util.Must(player.UpdateFromJson(context, data))
}

func FetchPlayer(context *util.Context) {
	// get user and player
	user := system.GetUser(context)
	player := GetPlayer(context)
	
	if player != nil {
		// add in user name and tag
		player.Name = user.Name
		player.Tag = user.Tag

		// update time sensetive player data
		util.Must(player.UpdateQuests(nil)) // should only write to the db once, so pass nil for context
		util.Must(player.UpdateTomes(context))
		
		// set all dirty flags
		player.SetAllDirty()
	} else {
		context.Fail(fmt.Sprintf("Failed to find player for user: %v", user.Name))
	}
}
