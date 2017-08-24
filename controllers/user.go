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
	handleGameAPI("/reauth", system.NoAuthentication, UserReauth)
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
	context.SetData("config", data.GameplayConfigJSON)
}

func UserLogin(context *util.Context) {
	// parse parameters
	reset := context.Params.GetBool("reset", false)
	development := context.Params.GetBool("development", false)

	// reset player data, if requested
	if reset {
		player, err := models.GetPlayerByUser(context, context.UserID)
		if player == nil {
			// create new player for user
			player, err = models.CreatePlayer(context.UserID, development)
			util.Must(err)

			util.Must(player.Save(context))
			
			// analytics tracking
			if util.HasSQLDatabase() {
				InsertTrackingSQL(context, "playerCreated", 0, "","", 0, 0, nil)
			}else{
				InsertTracking(context, "playerCreated", nil, 0)
			}
		} else {
			util.Must(err)

			// clear user name
			user := system.GetUser(context)
			user.Name = ""
			util.Must(user.Save(context))

			util.Must(player.Reset(context, development))
			
			// analytics tracking
			if util.HasSQLDatabase() {
				InsertTrackingSQL(context, "playerReset", 0, "","", 0, 0, nil)
			}else{
				InsertTracking(context, "playerReset", nil, 0)
			}
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
	valid, err := system.ValidateTokenWithoutClaims(context)
	util.Must(err)

	if valid {
		// issue new auth token
		system.IssueAuthToken(context)
	} else {
		context.Fail("Invalid token for reauthentication")
	}
}

func UserLogout(context *util.Context) {
	// clear auth token
	system.ClearAuthToken(context)
}
