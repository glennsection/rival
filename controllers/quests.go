package controllers

import (
	"strconv"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleQuests() {
	handleGameAPI("/quests/complete", system.TokenAuthentication, CompleteQuest)
	handleGameAPI("/quests/refresh", system.TokenAuthentication, RefreshQuests)
}

func CompleteQuest(context *util.Context) {
	player := GetPlayer(context)
	index, err := strconv.Atoi(context.Params.GetRequiredString("index"))

	if err != nil {
		panic(err)
	} else {
		if index < 0 || index > len(player.QuestSlots) {
			context.Fail("Invalid Index")
			return
		}
	}

	player.UpdateQuests(context)

	reward, success := player.CollectQuest(index, context)
	if !success {
		player.SetDirty(models.PlayerDataMask_Quests)
		context.Fail("Invalid Request")
		return
	}

	player.StandardCurrency += reward.StandardCurrency
	player.Save(context)

	player.SetDirty(models.PlayerDataMask_Quests, models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_XP)
	context.SetData("reward", reward)
}

func RefreshQuests(context *util.Context) {
	player := GetPlayer(context)

	if len(player.QuestSlots) < 3 {
		player.SetupQuestDefaults()
		player.Save(context)
	} else {
		player.UpdateQuests(context)
	}

	player.SetDirty(models.PlayerDataMask_Quests)
}