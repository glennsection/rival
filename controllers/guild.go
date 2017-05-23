package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func HandleGuild() {
	HandleGameAPI("/guild/create", system.TokenAuthentication, CreateGuild)
}

func CreateGuild(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get player
	player := GetPlayer(context)

	// create guild
	guild := models.CreateGuild(player.ID, name)
	util.Must(guild.Save(context.DB))

	// set guild for player
	player.GuildID = guild.ID
	util.Must(player.Save(context.DB))

	// set dirty for return data
	player.SetDirty(models.PlayerDataMask_Guild)
}