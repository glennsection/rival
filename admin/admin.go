package admin

import (
	"strings"

	"bloodtales/system"
)

var DefaultPageSize int = 20

func HandleAdmin(application *system.Application) {
	application.Redirect("/admin", "/admin/home", 301)
	handleAdminTemplate(application, "/admin/home", system.NoAuthentication, Home, "home.tmpl.html")
	handleAdminTemplate(application, "/admin/login", system.NoAuthentication, Login, "login.tmpl.html")
	handleAdminTemplate(application, "/admin/login/go", system.PasswordAuthentication, Login, "login.tmpl.html")
	handleAdminTemplate(application, "/admin/logout", system.NoAuthentication, Logout, "")
	handleAdminTemplate(application, "/admin/dashboard", system.TokenAuthentication, Dashboard, "dashboard.tmpl.html")

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
			URL: "/admin/dashboard",
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
			Name: "Matches",
			URL: "/admin/matches",
			Icon: "pe-7s-joy",
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
	if context.Authenticated() {
		context.Redirect("/admin/dashboard", 302)
	}
}

func Dashboard(context *system.Context) {
}

func Login(context *system.Context) {
	// handle request method
	switch context.Request.Method {
	case "POST":
		if context.Success {
			context.Message("User logged in successfully")
			context.Redirect("/admin/dashboard", 302)
		} else {
			context.Redirect("/admin/login", 302)
		}
	}
}

func Logout(context *system.Context) {
	// clear auth token
	context.ClearAuthentication()

	context.Message("User logged out successfully")
	context.Redirect("/admin", 302)
}
