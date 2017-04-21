package controllers

import (
	"fmt"
	
	"bloodtales/system"
	"bloodtales/models"
)

func HandleUser(application *system.Application) {
	application.HandleAPI("/register", system.NoAuthentication, UserRegister)
	application.HandleAPI("/login", system.PasswordAuthentication, UserLogin)
	application.HandleAPI("/logout", system.TokenAuthentication, UserLogout)
	//application.HandleAPI("/user/get", GetUser)
}

func UserRegister(context *system.Context) {
	// parse parameters
	username, password := context.Params.GetRequiredString("username"), context.Params.GetRequiredString("password")

	// insert user
	user, err := models.InsertUser(context.DB, username, password, false)
	if err != nil {
		panic(err)
	}
	context.User = user
	context.AppendToken()

	// create player
	player := models.CreatePlayer(user.ID, user.Username)
	err = player.Update(context.DB)
	if err != nil {
		panic(err)
	}

	if context.Success {
		context.Message("User registered successfully")
	}
}

func UserLogin(context *system.Context) {
	if context.Success {
		// analytics tracking (TODO - integrate with context)
		//context.Track("Login", bson.M { "mood": "happy" })
		
		context.Message("User logged in successfully")
	}
}

func UserLogout(context *system.Context) {
	if context.Success {
		context.Message("User logged out successfully")
	}
}

func GetUser(context *system.Context) {
	// parse parameters
	username := context.Params.GetRequiredString("username")

	// get user
	user, _ := models.GetUserByUsername(context.DB, username)
	if user != nil {
		context.Messagef("Found user: %v", user.Username)
	} else {
		panic(fmt.Sprintf("Failed to find User with username: %v", username))
	}
}
