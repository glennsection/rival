package controllers

import (
	"bloodtales/config"
	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
	"bloodtales/data"
)

func handleUser() {
	handleGameAPI("/connect", system.NoAuthentication, UserConnect)
	handleGameAPI("/login", system.DeviceAuthentication, UserLogin)
	handleGameAPI("/reauth", system.TokenAuthentication, UserReauth)
	handleGameAPI("/logout", system.TokenAuthentication, UserLogout)
}

func UserConnect(context *util.Context) {
	// parse parameters
	version := context.Params.GetRequiredString("version")

	// check version (major and minor)
	compatibility := util.CompareVersion(config.Config.Platform.Version, version, 2)
	switch compatibility {

	case -1:
		context.Fail("Client version is behind server.  Please update client!")

	case 1:
		context.Fail("Client version is ahead of server.  Please update server!")

	}

	// update client values
	context.Client.Version = version
	context.Client.Save()
	context.SetData("configuration", data.Config())
}

func UserLogin(context *util.Context) {
	// parse parameters
	reset := context.Params.GetBool("reset", false)

	// reset player data, if requested
	if reset {
		player, err := models.GetPlayerByUser(context, context.UserID)
		if player == nil {
			// create new player for user
			player, err = models.CreatePlayer(context.UserID)
			util.Must(err)

			util.Must(player.Save(context))
		} else {
			util.Must(err)

			util.Must(player.Reset(context))
		}
	}

	// analytics tracking
	if util.HasSQLDatabase() {
		InsertTrackingSQL(context, "login", 0, "","", 0, 0, nil)
	}else{
		InsertTracking(context, "login", nil, 0)
	}

	// respond with player data
	FetchPlayer(context)
}

func UserReauth(context *util.Context) {
	// issue new auth token
	system.IssueAuthToken(context)
}

func UserLogout(context *util.Context) {
	// clear auth token
	system.ClearAuthToken(context)
}
