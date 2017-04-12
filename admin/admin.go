package admin

import (
	"bloodtales/system"
)

var DefaultPageSize int = 20

func HandleAdmin(application *system.Application) {
	application.HandleTemplate("/admin", system.NoAuthentication, Home, "dashboard.tmpl.html")
	application.HandleTemplate("/admin/login", system.PasswordAuthentication, Login, "dashboard.tmpl.html")
	application.HandleTemplate("/admin/logout", system.NoAuthentication, Logout, "dashboard.tmpl.html")

	application.HandleTemplate("/admin/players", system.TokenAuthentication, ShowPlayers, "players.tmpl.html")
	application.HandleTemplate("/admin/player", system.TokenAuthentication, ShowPlayer, "player.tmpl.html")
}

func Home(session *system.Session) {
}

func Login(session *system.Session) {
	session.Message("User logged in successfully")
}

func Logout(session *system.Session) {
	// TODO - clear cookie?

	session.Message("User logged out successfully")
}
