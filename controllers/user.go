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
	username, password := context.GetRequiredParameter("username"), context.GetRequiredParameter("password")

	// insert user
	if err := models.InsertUser(context.Application.DB, username, password, false); err != nil {
		panic(err)
	}

	context.Message("User registered successfully")
}

func UserLogin(context *system.Context) {
	context.Message("User logged in successfully")
}

func UserLogout(context *system.Context) {
	// TODO - clear token?

	context.Message("User logged out successfully")
}

func GetUser(context *system.Context) {
	// parse parameters
	username := context.GetRequiredParameter("username")

	// get user
	user, _ := models.GetUserByUsername(context.Application.DB, username)
	if user != nil {
		context.Messagef("Found user: %v", user.Username)
	} else {
		panic(fmt.Sprintf("Failed to find User with username: %v", username))
	}
}
