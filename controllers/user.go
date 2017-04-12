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

func UserRegister(session *system.Session) {
	// parse parameters
	username, password := session.GetRequiredParameter("username"), session.GetRequiredParameter("password")

	// insert user
	if err := models.InsertUser(session.Application.DB, username, password, false); err != nil {
		panic(err)
	}

	session.Message("User registered successfully")
}

func UserLogin(session *system.Session) {
	session.Message("User logged in successfully")
}

func UserLogout(session *system.Session) {
	// TODO - clear token?

	session.Message("User logged out successfully")
}

func GetUser(session *system.Session) {
	// parse parameters
	username := session.GetRequiredParameter("username")

	// get user
	user, _ := models.GetUserByUsername(session.Application.DB, username)
	if user != nil {
		session.Messagef("Found user: %v", user.Username)
	} else {
		panic(fmt.Sprintf("Failed to find User with username: %v", username))
	}
}
