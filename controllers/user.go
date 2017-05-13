package controllers

import (
	"bloodtales/system"
	"bloodtales/util"
)

func HandleUser(application *system.Application) {
	HandleGameAPI(application, "/connect", system.NoAuthentication, UserConnect)
	//HandleGameAPI(application, "/register", system.NoAuthentication, UserRegister)
	HandleGameAPI(application, "/login", system.DeviceAuthentication, UserLogin)
	HandleGameAPI(application, "/logout", system.TokenAuthentication, UserLogout)
}

func UserConnect(context *system.Context) {
	// parse parameters
	version := context.Params.GetRequiredString("version")

	// check version (major and minor)
	compatibility := util.CompareVersion(context.Config.Platform.Version, version, 2)
	switch compatibility {

	case -1:
		context.Fail("Client version is behind server.  Please update client!")

	case 1:
		context.Fail("Client version is ahead of server.  Please update server!")

	}

	// update client values
	context.Client.Version = version
	context.Client.Save()
}

func UserLogin(context *system.Context) {
	if context.Success {
		// analytics tracking (TODO - integrate with context)
		//context.Track("Login", bson.M { "mood": "happy" })

		// respond with player data
		FetchPlayer(context)
	}
}

func UserLogout(context *system.Context) {
	// clear auth token
	context.ClearAuthToken()

	if context.Success {
		context.Message("User logged out successfully")
	}
}
