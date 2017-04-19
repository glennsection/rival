package admin

import (
	"strings"

	"bloodtales/system"
)

var DefaultPageSize int = 20

func HandleAdmin(application *system.Application) {
	application.Redirect("/admin", "/admin/home", 301)
	handleAdminTemplate(application, "/admin/home", system.NoAuthentication, Home, "dashboard.tmpl.html")
	handleAdminTemplate(application, "/admin/login", system.PasswordAuthentication, Login, "dashboard.tmpl.html")
	handleAdminTemplate(application, "/admin/logout", system.NoAuthentication, Logout, "dashboard.tmpl.html")

	handleAdminUsers(application)
	handleAdminAnalytics(application)
}

func handleAdminTemplate(application *system.Application, pattern string, authType system.AuthenticationType, handler func(*system.Context), template string) {
	application.HandleTemplate(pattern, authType, func(context *system.Context) {
		initializeAdmin(context)
		handler(context)
	}, template)
}

func initializeAdmin(context *system.Context) {
	// sidebar links
	links := []struct {
		Name string
		URL string
		Icon string
		Active bool
	} {
		{
			Name: "Dashboard",
			URL: "/admin/home",
			Icon: "pe-7s-graph2",
		},
		{
			Name: "Players",
			URL: "/admin/users",
			Icon: "pe-7s-users",
		},
		{
			Name: "Leaderboard",
			URL: "/admin/leaderboard",
			Icon: "pe-7s-cup",
		},
		{
			Name: "Events",
			URL: "/admin/events",
			Icon: "pe-7s-timer",
		},
		{
			Name: "Content",
			URL: "/admin/content",
			Icon: "pe-7s-gift",
		},
	}

	for i, link := range links {
		if strings.HasPrefix(context.Request.URL.Path, link.URL) {
			links[i].Active = true
		}
	}

	context.Params.Set("links", links)
}

func Home(context *system.Context) {
}

func Login(context *system.Context) {
	context.Message("User logged in successfully")
}

func Logout(context *system.Context) {
	// TODO - clear cookie?
	context.Message("User logged out successfully")
}
