package main

import (
	"fmt"
	"net/http"

	"bloodtales/system"
	"bloodtales/controllers"
)

func main() {
	// init application
	var application = &system.Application{}
	defer application.Close()
	application.Init()

	// route request
	application.Handle("/", root)
	application.Handle("/user/new", controllers.InsertUser)
	application.Handle("/user/get", controllers.GetUser)
	application.Handle("/user/login", controllers.LoginUser)

	// deliver response
	application.Serve()
}

func root(w http.ResponseWriter, r *http.Request, app *system.Application) {
	fmt.Fprint(w, "Invalid request")
}
