package controllers

import (
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/data"
	"bloodtales/models"
)

func handleQuests() {
	handleGameAPI("/quests/complete", system.TokenAuthentication, CompleteQuest)
	handleGameAPI("/quests/clear", system.TokenAuthentication, ClearQuest)
	handleGameAPI("/quests/refresh", system.TokenAuthentication, RefreshQuests)
}

func CompleteQuest(context *util.Context) {
	player, index, valid := validateQuestRequest(context)
	if !valid {
		return
	}

	player.UpdateQuests(context)

	questId := player.QuestSlots[index].QuestInstance.DataID //cache for analytics

	reward, success, err := player.CollectQuest(index, context)
	util.Must(err)

	if !success {
		player.SetDirty(models.PlayerDataMask_Quests)
		context.Fail("Quest not ready to be collected")
		return
	}

	player.SetDirty(models.PlayerDataMask_Quests, models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_XP)
	context.SetData("reward", reward)

	currentTime := util.TimeToTicks(time.Now().UTC())

	if util.HasSQLDatabase() {
		InsertTrackingSQL(context, "questComplete", currentTime, data.ToDataName(data.GetQuestData(questId).RewardID), data.ToDataName(questId), 1, 0, nil)

		TrackRewardsSQL(context, reward, currentTime)	
	} else{
		InsertTracking(context, "questComplete", bson.M { "time":currentTime,
													  "questId":data.ToDataName(questId),
													  "rewardId":data.ToDataName(data.GetQuestData(questId).RewardID) }, 0)

		TrackRewards(context, reward)	
	}
}

func ClearQuest(context *util.Context) {
	player, index, valid := validateQuestRequest(context)
	if !valid {
		return
	}

	player.SetDirty(models.PlayerDataMask_Quests)

	if player.QuestClearTime > util.TimeToTicks(time.Now().UTC()) {
		context.Fail("Cannot clear any quests at this time")
		return
	}

	if player.QuestSlots[index].State == models.QuestState_Ready || player.QuestSlots[index].State == models.QuestState_Cooldown {
		context.Fail("Cannot clear quest at this time")
		return
	}

	player.QuestSlots[index].State = models.QuestState_Ready
	player.AssignRandomQuest(index)
	player.QuestClearTime = util.TimeToTicks(util.GetTomorrowDate())
	player.Save(context)
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

func validateQuestRequest(context *util.Context) (*models.Player, int, bool) {
	success := true
	player := GetPlayer(context)
	index, err := strconv.Atoi(context.Params.GetRequiredString("index"))
	util.Must(err)

	if index < 0 || index > len(player.QuestSlots) {
		success = false
		context.Fail("Invalid Index")
	}

	return player, index, success
}