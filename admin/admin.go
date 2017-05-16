package admin

import (
	"strings"

	"bloodtales/system"
)

var DefaultPageSize int = 20

func HandleAdmin() {
	system.App.Redirect("/admin", "/admin/home", 301)
	handleAdminTemplate("/error", system.NoAuthentication, Error, "error.tmpl.html")
	handleAdminTemplate("/admin/home", system.NoAuthentication, Home, "home.tmpl.html")
	handleAdminTemplate("/admin/login", system.NoAuthentication, Login, "login.tmpl.html")
	handleAdminTemplate("/admin/login/go", system.PasswordAuthentication, Login, "login.tmpl.html")
	handleAdminTemplate("/admin/logout", system.NoAuthentication, Logout, "")
	handleAdminTemplate("/admin/dashboard", system.TokenAuthentication, Dashboard, "dashboard.tmpl.html")

	handleAdminUsers()
	handleAdminCards()
	handleAdminAnalytics()
}

func handleAdminTemplate(pattern string, authType system.AuthenticationType, handler func(*system.Context), template string) {
	system.App.HandleTemplate(pattern, authType, func(context *system.Context) {
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

func Error(context *system.Context) {
	// parse parameters
	message := context.Params.GetString("message", "Error occurred")

	context.Message(message) // TODO - fix this once session flashes are working
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
			user := system.GetUser(context)

			if user.Admin == true {
				context.Message("User logged in successfully")
				context.Redirect("/admin/dashboard", 302)
			} else {
				context.Fail("User is not an admin")
				context.Redirect("/admin/login", 302)
			}
		} else {
			context.Redirect("/admin/login", 302)
		}
	}
}

func Logout(context *system.Context) {
	// clear auth token
	context.ClearAuthToken()

	context.Message("User logged out successfully")
	context.Redirect("/admin", 302)
}
