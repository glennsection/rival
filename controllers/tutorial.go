package controllers

import (
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleTutorial() {
	handleGameAPI("/tutorial/claimTome", system.TokenAuthentication, ClaimTutorialTome)
}

func ClaimTutorialTome(context *util.Context) {
	player := GetPlayer(context)

	// add the tutorial tome to the player's inventory
	player.ClaimTutorialTome(context)

	player.SetDirty(models.PlayerDataMask_Tomes)
}