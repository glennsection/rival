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
	system.AddTemplateFunc("getUserName", templateGetUserName)
	system.AddTemplateFunc("getPlayerName", templateGetPlayerName)
}

func GetPlayer(context *system.Context) (player *models.Player) {
	player, _ = models.GetPlayerByUser(context.DB, context.User.ID)
	return
}

func RefreshUserName(context *system.Context, name string, userID bson.ObjectId, playerID bson.ObjectId) {
	userKey := fmt.Sprintf("UserName:%s", userID.Hex())
	playerKey := fmt.Sprintf("UserPlayerName:%s", playerID.Hex())

	context.Cache.Set(userKey, name)
	context.Cache.Set(playerKey, name)
}

func templateGetUserName(context *system.Context, userID bson.ObjectId) string {
	key := fmt.Sprintf("UserName:%s", userID.Hex())
	name := ""

	if context.Cache.Has(key) {
		name = context.Cache.GetString(key, "")
	}

	if name == "" {
		user, err := models.GetUserById(context.DB, userID)
		if err == nil && user != nil {
			context.Cache.Set(key, user.Name)
			name = user.Name
		}
	}
	return name
}

func templateGetPlayerName(context *system.Context, playerID bson.ObjectId) string {
	key := fmt.Sprintf("UserPlayerName:%s", playerID.Hex())
	name := ""

	if context.Cache.Has(key) {
		name = context.Cache.GetString(key, "")
	}

	// if name == "" { // TODO FIXME...
	// 	user, err := models.GetUserByPlayerID(context.DB, playerID)
	// 	if err == nil && user != nil {
	// 		context.Cache.Set(key, user.Name)
	// 		name = user.Name
	// 	}
	// }
	return name
}

func SetPlayerName(context *system.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get user
	user := context.User

	// set name and update
	user.Name = name
	err := user.Update(context.DB)
	if err != nil {
		panic(err)
	}

	// get player
	player, err := models.GetPlayerByUser(context.DB, user.ID)
	if err != nil {
		panic(err)
	}

	// refresh cached name
	RefreshUserName(context, name, user.ID, player.ID)
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
		// update rewards
		err := player.UpdateRewards(context.DB)
		if(err != nil) {
			panic(err)
		}

		// add in user name
		player.Name = context.User.Name
		
		// set successful response
		context.Message("Found player")
		context.Data = player
	} else {
		context.Fail(fmt.Sprintf("Failed to find player for username: %v", context.User.Username))
	}
}
