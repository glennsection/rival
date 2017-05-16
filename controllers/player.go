package controllers

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func HandlePlayer() {
	HandleGameAPI("/player/set", system.TokenAuthentication, SetPlayer)
	HandleGameAPI("/player/name", system.TokenAuthentication, SetPlayerName)
	//HandleGameAPI("/player/get", system.TokenAuthentication, GetPlayer)
	HandleGameAPI("/player/refresh", system.TokenAuthentication, updateAllPlayersPlace)

	// template functions
	util.AddTemplateFunc("getUserName", templateGetUserName)
	util.AddTemplateFunc("getPlayerName", templateGetPlayerName)
}

func GetPlayer(context *system.Context) (player *models.Player) {
	player, ok := context.Params.Get("_player").(*models.Player)
	if ok == false {
		user := system.GetUser(context)
		player, _ = models.GetPlayerByUser(context.DB, user.ID)

		if player != nil {
			context.Params.Set("_player", player)
		}
	}
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

	name := context.Cache.GetString(key, "")

	// immediately cache latest name
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

	name := context.Cache.GetString(key, "")

	// immediately cache latest name
	if name == "" {
		player, _ := models.GetPlayerById(context.DB, playerID)
		if player != nil {
			user, _ := models.GetUserById(context.DB, player.UserID)
			if user != nil {
				context.Cache.Set(key, user.Name)
				name = user.Name
			}
		}
	}
	return name
}

func templateGetPlayerPlace(context *system.Context, player *models.Player) int {
	return 0;
	// key := fmt.Sprintf("UserName:%s", userID.Hex())

	// name := context.Cache.GetString(key, "")

	// // immediately cache latest name
	// if name == "" {
	// 	user, err := models.GetUserById(context.DB, userID)
	// 	if err == nil && user != nil {
	// 		context.Cache.Set(key, user.Name)
	// 		name = user.Name
	// 	}
	// }
	// return name
}

func updateAllPlayersPlace(context *system.Context) {
	var players []*models.Player
	context.DB.C(models.PlayerCollectionName).Find(nil).All(&players)

	for _, player := range players {
		updatePlayerPlace(context, player)
	}
}

func updatePlayerPlace(context *system.Context, player *models.Player) {
	matches := player.MatchCount
	if matches > 0 {
		// calculate placement score
		winsFactor := player.WinCount * 1000000 / matches
		matchesFactor := matches * 1000
		pointsFactor := player.ArenaPoints

		score := winsFactor + matchesFactor + pointsFactor
		context.Cache.SetScore("Leaderboard", player.ID.Hex(), score)
	}
}

func SetPlayerName(context *system.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get user
	user := system.GetUser(context)

	// set name and update
	user.Name = name
	err := user.Save(context.DB)
	util.Must(err)

	// get player
	player, err := models.GetPlayerByUser(context.DB, user.ID)
	util.Must(err)

	// refresh cached name
	RefreshUserName(context, name, user.ID, player.ID)
}

func SetPlayer(context *system.Context) {
	// parse parameters
	data := context.Params.GetRequiredString("data")

	// update data
	user := system.GetUser(context)
	_, err := models.UpdatePlayer(context.DB, user, data)
	util.Must(err)

	context.Message("Player updated successfully")
}

func FetchPlayer(context *system.Context) {
	// get user and player
	user := system.GetUser(context)
	player := GetPlayer(context)
	
	if player != nil {
		// update rewards
		util.Must(player.UpdateRewards(context.DB))

		// add in user name
		player.Name = user.Name
		
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
		context.Fail(fmt.Sprintf("Failed to find player for username: %v", user.Username))
	}
}
