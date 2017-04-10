package admin

import (
	//"log"

	"bloodtales/system"
	//"bloodtales/models"
)

func Home(session *system.Session) {
	session.Template = "index.tmpl.html"
}

func Login(session *system.Session) {
	session.Message("User logged in successfully")
}

func Logout(session *system.Session) {
	// TODO - clear token?

	session.Message("User logged out successfully")
}
