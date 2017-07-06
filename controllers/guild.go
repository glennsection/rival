package controllers

import (
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"

	"gopkg.in/mgo.v2/bson"

	"fmt"
)

func handleGuild() {
	handleGameAPI("/guild/create", system.TokenAuthentication, CreateGuild)
	handleGameAPI("/guild/getGuilds", system.TokenAuthentication, GetGuilds)
	handleGameAPI("/guild/addMember", system.TokenAuthentication, AddMember)
}

func CreateGuild(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get player
	player := GetPlayer(context)

	// TODO - make sure player doesn't already own a guild...

	// create guild
	_, err := models.CreateGuild(context, player, name)
	util.Must(err)
}

func GetGuilds(context *util.Context) {
	// parse parameters
	name := context.Params.GetString("name", "")

	var guilds []*models.Guild
	if name != "" {
		util.Must(context.DB.C(models.GuildCollectionName).Find(bson.M{
			"nm": bson.M{
				"$regex": bson.RegEx{
					Pattern: fmt.Sprintf(".*%s.*", name),
					Options: "i",
				},
			},
		}).All(&guilds))
	} else {
		util.Must(context.DB.C(models.GuildCollectionName).Find(nil).All(&guilds))
	}
	//util.Must(context.DB.C(models.GuildCollectionName).Find(bson.M{"nm": bson.M{"$eq": "MikeMasterGuild"}}).All(&guilds))

	// result
	context.SetData("guilds", guilds)
}

func AddMember(context *util.Context) {
	tag := context.Params.GetRequiredString("tag")

	player := GetPlayer(context)

	// guild
	guild, err := models.GetGuildByTag(context, tag)
	util.Must(err)

	err2 := models.AddMember(context, player, guild)
	util.Must(err2)
}
