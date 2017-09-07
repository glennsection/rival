package main

import (
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/controllers"
	"bloodtales/admin"
	_ "bloodtales/testing"
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

func root(context *util.Context) {
	// root is invalid for now
	context.Fail("Invalid request")
}
