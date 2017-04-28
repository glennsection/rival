package controllers

import (
	// "fmt"
	
	"bloodtales/system"
	// "bloodtales/models"
)

func HandleUser(application *system.Application) {
	application.HandleAPI("/connect", system.NoAuthentication, UserConnect)
	//application.HandleAPI("/register", system.NoAuthentication, UserRegister)
	application.HandleAPI("/login", system.DeviceAuthentication, UserLogin)
	application.HandleAPI("/logout", system.TokenAuthentication, UserLogout)
	//application.HandleAPI("/user/get", GetUser)
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
		GetPlayer(context)
	}
}

func UserLogout(context *system.Context) {
	// clear auth token
	context.ClearAuthToken()

	if context.Success {
		context.Message("User logged out successfully")
	}
}

// func GetUser(context *system.Context) {
// 	// parse parameters
// 	username := context.Params.GetRequiredString("username")

// 	// get user
// 	user, _ := models.GetUserByUsername(context.DB, username)
// 	if user != nil {
// 		context.Messagef("Found user: %v", user.Username)
// 	} else {
// 		context.Fail(fmt.Sprintf("Failed to find User with username: %v", username))
// 	}
// }
