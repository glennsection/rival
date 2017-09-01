package controllers

import (
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
)

func handleTutorial() {
	handleGameAPI("/tutorial/updateProgress", system.TokenAuthentication, UpdateTutorialProgress)
	handleGameAPI("/tutorial/claimReward", system.TokenAuthentication, ClaimTutorialReward)
}

func UpdateTutorialProgress(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")
	complete := context.Params.GetRequiredBool("complete")
	page := context.Params.GetRequiredInt("page")
	progress := context.Params.GetRequiredInt("progress")

	// get player
	player := GetPlayer(context)

	//validate params
	//TODO

	err := models.UpdateTutorial(context, player, name, complete, page, progress)
	util.Must(err)
}

func ClaimTutorialReward(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get player
	player := GetPlayer(context)

	tome, reward, err := player.ClaimTutorialReward(context, name)
	util.Must(err)

	if tome != nil {
		player.SetDirty(models.PlayerDataMask_Tomes)
		context.SetData("tome", tome)
	}

	if reward != nil {
		player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards)
		context.SetData("reward", reward)
	}

	player.SetDirty(models.PlayerDataMask_Stars)
}