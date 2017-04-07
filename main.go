package main

import (
	"bloodtales/system"
	"bloodtales/controllers"
)

func main() {
	// init application
	var application = &system.Application {}
	defer application.Close()
	application.Init()

	// -------- Routes --------

	application.Handle("/", system.NoAuthentication, root)

	application.Handle("/register", system.NoAuthentication, controllers.RegisterUser)
	application.Handle("/login", system.PasswordAuthentication, controllers.LoginUser)
	application.Handle("/logout", system.TokenAuthentication, controllers.LogoutUser)
	//application.Handle("/user/get", controllers.GetUser)

	application.Handle("/player/set", system.TokenAuthentication, controllers.SetPlayer)
	application.Handle("/player/get", system.TokenAuthentication, controllers.GetPlayer)

	// ------------------------

	// deliver response
	application.Serve()
}

func root(session *system.Session) {
	// root is invalid for now
	session.Fail("Invalid request")
}
