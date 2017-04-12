package main

import (
	"bloodtales/system"
	"bloodtales/controllers"
	"bloodtales/admin"
)

func main() {
	// init application
	var application = &system.Application {}
	defer application.Close()
	application.Initialize()

	// -------- Routes --------

	application.Ignore("/")
	application.Ignore("/favicon.ico")

	application.Handle("/admin", system.NoAuthentication, admin.Home)

	application.Handle("/register", system.NoAuthentication, controllers.UserRegister)
	application.Handle("/login", system.PasswordAuthentication, controllers.UserLogin)
	application.Handle("/logout", system.TokenAuthentication, controllers.UserLogout)
	//application.Handle("/user/get", controllers.GetUser)

	application.Handle("/player/set", system.TokenAuthentication, controllers.SetPlayer)
	application.Handle("/player/get", system.TokenAuthentication, controllers.GetPlayer)

	application.Handle("/player/tome/unlock", system.TokenAuthentication, controllers.UnlockTome)
	application.Handle("/player/tome/open", system.TokenAuthentication, controllers.OpenTome)
	// ------------------------

	// deliver response
	application.Serve()
}

func root(session *system.Session) {
	// root is invalid for now
	session.Fail("Invalid request")
}
