package controllers

import (
	"bloodtales/system"
)

func HandleUser(application *system.Application) {
	application.HandleAPI("/connect", system.NoAuthentication, UserConnect)
	//application.HandleAPI("/register", system.NoAuthentication, UserRegister)
	application.HandleAPI("/login", system.DeviceAuthentication, UserLogin)
	application.HandleAPI("/logout", system.TokenAuthentication, UserLogout)
}

func UserConnect(context *system.Context) {
	// parse parameters
	version := context.Params.GetRequiredString("version")

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
