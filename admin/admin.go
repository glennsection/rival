package admin

import (
	//"log"

	"bloodtales/system"
	//"bloodtales/models"
)

func HandleAdmin(application *system.Application) {
	application.Handle("/admin", system.NoAuthentication, Home)
}

func Home(session *system.Session) {
	session.Template = "dashboard.tmpl.html"
}

func Login(session *system.Session) {
	session.Message("User logged in successfully")
}

func Logout(session *system.Session) {
	// TODO - clear token?

	session.Message("User logged out successfully")
}
