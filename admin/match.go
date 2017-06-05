package admin

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminMatches() {
	handleAdminTemplate("/admin/matches", system.TokenAuthentication, ViewMatches, "matches.tmpl.html")
	handleAdminTemplate("/admin/matches/edit", system.TokenAuthentication, EditMatch, "match.tmpl.html")
	handleAdminTemplate("/admin/matches/delete", system.TokenAuthentication, DeleteMatch, "")
}

func ViewMatches(context *util.Context) {
	// paginate players query
	pagination, err := context.Paginate(context.DB.C(models.MatchCollectionName).Find(nil).Sort("-t0"), DefaultPageSize)
	util.Must(err)

	// get resulting matches
	var matches []*models.Match
	util.Must(pagination.All(&matches))

	// set template bindings
	context.Params.Set("matches", matches)
}

func EditMatch(context *util.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredId("matchId")

	match, err := models.GetMatchById(context, matchId)
	util.Must(err)

	// handle request method
	switch context.Request.Method {
	case "POST":
		context.Message("Match updated!")
	}
	
	// set template bindings
	context.Params.Set("match", match)
}

func DeleteMatch(context *util.Context) {
	// parse parameters
	matchId := context.Params.GetRequiredId("matchId")
	page := context.Params.GetInt("page", 1)

	match, err := models.GetMatchById(context, matchId)
	util.Must(err)

	match.Delete(context)

	context.Redirect(fmt.Sprintf("/admin/matches?page=%d", page), 302)
}
