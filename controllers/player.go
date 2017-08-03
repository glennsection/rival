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
	handleGameAPI("/player/view", system.NoAuthentication, ViewPlayerProfile)

	// template functions
	util.AddTemplateFunc("getUserName", models.GetUserName)
	util.AddTemplateFunc("getPlayerName", models.GetPlayerName)
}

func GetPlayer(context *util.Context) (player *models.Player) {
	// get player for current context, with cache in params
	player, ok := context.Params.Get("_player").(*models.Player)
	if ok == false {
		player, _ = models.GetPlayerByUser(context, context.UserID)

		if player != nil {
			context.Params.Set("_player", player)
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

func ViewPlayerProfile(context *util.Context) {
	tag := context.Params.GetRequiredString("tag")

	var player *models.Player
	var err error

	if player, err = models.GetPlayerByTag(context, tag); err != nil || player == nil {
		panic(err)
	}

	context.SetData("name", player.Name)
	context.SetData("rankPoints", player.RankPoints)
	context.SetData("wins", player.WinCount)
	context.SetData("cards", len(player.Cards))
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
		context.Fail(fmt.Sprintf("Failed to find player for user: %v (%v)", user.Name, user.ID.Hex()))
	}
}
