package controllers

import (
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
)

func handleTutorial() {
	handleGameAPI("/tutorial/updateProgress", system.TokenAuthentication, UpdateTutorialProgress)
}

func UpdateTutorialProgress(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")
	complete := context.Params.GetRequiredBool("complete")

	// get player
	player := GetPlayer(context)

	//validate params
	//TODO

	err := models.UpdateTutorial(context, player, name, complete)
	util.Must(err)
}
