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

	// TODO - make sure player doesn't already own a guild...

	// create guild
	_, err := models.CreateGuild(context.DB, player, name)
	util.Must(err)
}

