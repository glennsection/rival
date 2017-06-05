package controllers

import (
	"strconv"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleQuests() {
	handleGameAPI("/quests/complete", system.TokenAuthentication, CompleteQuest)
}

func CompleteQuest(context *util.Context) {
	player := GetPlayer(context)
	index, err := strconv.Atoi(context.Params.GetRequiredString("index"))

	if err != nil {
		panic(err)
	} else {
		if index < 0 || index > len(player.Quests) {
			context.Fail("Invalid Index")
		}
	}

	player.UpdateQuests()

	reward, success := player.CollectQuest(index, context.DB)
	if !success {
		context.Fail("Invalid Request")
	}

	player.SetDirty(models.PlayerDataMask_Quests)
	context.SetData("reward", reward)
}