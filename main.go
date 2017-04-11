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
	application.Static("/static", "static")

	admin.HandleAdmin(application)

	controllers.HandleUser(application)
	controllers.HandlePlayer(application)
	controllers.HandleTome(application)

	// ------------------------

	// deliver response
	application.Serve()
}

func root(session *system.Session) {
	// root is invalid for now
	session.Fail("Invalid request")
}
