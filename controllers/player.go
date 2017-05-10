package controllers

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
)

func HandlePlayer(application *system.Application) {
	application.HandleAPI("/player/set", system.TokenAuthentication, SetPlayer)
	application.HandleAPI("/player/name", system.TokenAuthentication, SetPlayerName)
	//application.HandleAPI("/player/get", system.TokenAuthentication, GetPlayer)

	// template functions
	system.AddTemplateFunc("getUserPlayerName", GetUserPlayerName)
	system.AddTemplateFunc("getPlayerName", GetPlayerName)
}

func GetPlayer(context *system.Context) (player *models.Player) {
	player, _ = models.GetPlayerByUser(context.DB, context.User.ID)
	return
}

func RefreshPlayerName(context *system.Context, player *models.Player) {
	playerKey := fmt.Sprintf("PlayerName:%s", player.ID.Hex())
	userKey := fmt.Sprintf("UserPlayerName:%s", player.UserID.Hex())

	context.Cache.Set(playerKey, player.Name)
	context.Cache.Set(userKey, player.Name)
}

func GetUserPlayerName(context *system.Context, userID bson.ObjectId) string {
	key := fmt.Sprintf("UserPlayerName:%s", userID.Hex())
	name := ""

	if context.Cache.Has(key) {
		name = context.Cache.GetString(key, "")
	}

	if name == "" {
		player, err := models.GetPlayerByUser(context.DB, userID)
		if err == nil && player != nil {
			context.Cache.Set(key, player.Name)
			name = player.Name
		}
	}
	return name
}

func GetPlayerName(context *system.Context, playerID bson.ObjectId) string {
	key := fmt.Sprintf("PlayerName:%s", playerID.Hex())
	name := ""

	if context.Cache.Has(key) {
		name = context.Cache.GetString(key, "")
	}

	if name == "" {
		player, err := models.GetPlayerById(context.DB, playerID)
		if err == nil && player != nil {
			context.Cache.Set(key, player.Name)
			name = player.Name
		}
	}
	return name
}

func SetPlayerName(context *system.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get player
	player := GetPlayer(context)

	// set name and update
	player.Name = name
	err := player.Update(context.DB)
	if err != nil {
		panic(err)
	}

	// refresh cached name
	RefreshPlayerName(context, player)
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

func FetchPlayer(context *system.Context) {
	// get player
	player := GetPlayer(context)
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
