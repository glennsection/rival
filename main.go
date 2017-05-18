package main

import (
	"bloodtales/system"
	"bloodtales/controllers"
	"bloodtales/admin"
)

func main() {
	// init application
	application := system.App
	defer application.Close()

	// -------- Routes --------
	application.Ignore("/")
	application.Ignore("/favicon.ico")
	application.Static("/static", "static")

	admin.HandleAdmin()
	controllers.HandleGame()
	// ------------------------

	// listen and serve responses
	application.Serve()
}

func root(context *system.Context) {
	// root is invalid for now
	context.Fail("Invalid request")
}
