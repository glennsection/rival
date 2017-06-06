package controllers

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handlePlayer() {
	handleGameAPI("/player/set", system.TokenAuthentication, OverwritePlayer) // HACK
	//handleGameAPI("/player/get", system.TokenAuthentication, GetPlayer)
	handleGameAPI("/player/name", system.TokenAuthentication, SetPlayerName)
	handleGameAPI("/player/refresh", system.TokenAuthentication, models.UpdateAllPlayersPlace) // HACK

	// template functions
	util.AddTemplateFunc("getUserName", GetUserName)
	util.AddTemplateFunc("getPlayerName", GetPlayerName)
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

	// get cache keys
	userKey := fmt.Sprintf("UserName:%s", user.ID.Hex())
	playerKey := fmt.Sprintf("UserPlayerName:%s", player.ID.Hex())

	// refresh cached names
	context.Cache.Set(userKey, name)
	context.Cache.Set(playerKey, name)
}

func GetUserName(context *util.Context, userID bson.ObjectId) string {
	// get cache key
	key := fmt.Sprintf("UserName:%s", userID.Hex())

	// get cached name
	name := context.Cache.GetString(key, "")

	// immediately cache latest name
	if name == "" {
		user, err := models.GetUserById(context, userID)
		if err == nil && user != nil {
			context.Cache.Set(key, user.Name)
			name = user.Name
		}
	}
	return name
}

func GetPlayerName(context *util.Context, playerID bson.ObjectId) string {
	// get cache key
	key := fmt.Sprintf("UserPlayerName:%s", playerID.Hex())

	// get cached name
	name := context.Cache.GetString(key, "")

	// immediately cache latest name
	if name == "" {
		player, _ := models.GetPlayerById(context, playerID)
		if player != nil {
			user, _ := models.GetUserById(context, player.UserID)
			if user != nil {
				context.Cache.Set(key, user.Name)
				name = user.Name
			}
		}
	}
	return name
}

func GetUserIdByPlayerId(context *util.Context, playerID bson.ObjectId) bson.ObjectId {
	// get cache key
	key := fmt.Sprintf("PlayerUserId:%s", playerID.Hex())

	// get cached ID
	userIDHex := context.Cache.GetString(key, "")
	var userID bson.ObjectId

	if bson.IsObjectIdHex(userIDHex) {
		// user cached ID
		userID = bson.ObjectIdHex(userIDHex)
	} else {
		// get and cache ID
		player, _ := models.GetPlayerById(context, playerID)
		if player != nil {
			userID = player.UserID
			context.Cache.Set(key, userID.Hex())
		}
	}
	return userID
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
		context.Fail(fmt.Sprintf("Failed to find player for username: %v", user.Username))
	}
}
