package admin

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminGuilds() {
	handleAdminTemplate("/admin/guilds", system.TokenAuthentication, ViewGuilds, "guilds.tmpl.html")
	handleAdminTemplate("/admin/guilds/edit", system.TokenAuthentication, EditGuild, "guild.tmpl.html")
	handleAdminTemplate("/admin/guilds/delete", system.TokenAuthentication, DeleteGuild, "")
}

func ViewGuilds(context *util.Context) {
	// parse parameters
	search := context.Params.GetString("search", "")

	// process search terms
	var query *mgo.Query = nil
	if search != "" {
		// build query
		query = context.DB.C(models.GuildCollectionName).Find(bson.M {
			"nm": bson.M {
				"$regex": bson.RegEx {
					Pattern: fmt.Sprintf(".*%s.*", search),
					Options: "i",
				},
			},
		})
	} else {
		query = context.DB.C(models.GuildCollectionName).Find(nil)
	}

	// sorting
	query = context.Sort(query, "")

	// paginate guilds query
	pagination, err := context.Paginate(query, DefaultPageSize)
	util.Must(err)

	// get resulting guilds
	var guilds []*models.Guild
	util.Must(pagination.All(&guilds))

	// set template bindings
	context.Params.Set("guilds", guilds)
}

func EditGuild(context *util.Context) {
	// parse parameters
	guildId := context.Params.GetRequiredId("guildId")

	guild, err := models.GetGuildById(context, guildId)
	util.Must(err)

	// get members
	var members []*models.Player
	util.Must(context.DB.C(models.PlayerCollectionName).Find(bson.M { "gd": guild.ID } ).All(&members))

	// handle request method
	switch context.Request.Method {
	case "POST":
		// TODO
		context.Message("Guild updated!")
	}
	
	// set template bindings
	context.Params.Set("guild", guild)
	context.Params.Set("members", members)
}

func DeleteGuild(context *util.Context) {
	// parse parameters
	guildId := context.Params.GetRequiredId("guildId")
	page := context.Params.GetInt("page", 1)

	guild, err := models.GetGuildById(context, guildId)
	util.Must(err)

	guild.Delete(context)

	context.Redirect(fmt.Sprintf("/admin/guilds?page=%d", page), 302)
}
