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

	questId := player.Quests[index].QuestID //cache for analytics

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
		InsertTrackingSQL(context, "questComplete", currentTime, reward.Data.ID, data.ToDataName(questId), 1, 0, nil)

		TrackRewardsSQL(context, reward, currentTime)	
	} else{
		InsertTracking(context, "questComplete", bson.M { "time":currentTime,
													  "questId":data.ToDataName(questId),
													  "rewardId":reward.Data.ID }, 0)

		TrackRewards(context, reward)	
	}
}

func ClearQuest(context *util.Context) {
	player, index, valid := validateQuestRequest(context)
	if !valid {
		return
	}

	player.SetDirty(models.PlayerDataMask_Quests)

	if !data.GetQuestData(player.Quests[index].QuestID).Disposable {
		context.Fail("This quest cannot be cleared")
		return
	}

	if player.QuestClearTime > util.TimeToTicks(time.Now().UTC()) {
		context.Fail("Cannot clear any quests at this time")
		return
	}

	if !player.Quests[index].Active {
		context.Fail("Cannot clear quest at this time")
		return
	}

	player.Quests[index].Active = false
	player.AssignRandomQuest(index)
	player.QuestClearTime = util.TimeToTicks(util.GetDateInNDays(player.TimeZone, 1)) // tomorrow
	player.Save(context)
}

func RefreshQuests(context *util.Context) {
	player := GetPlayer(context)

	if len(player.Quests) < 3 {
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

	if index < 0 || index > len(player.Quests) {
		success = false
		context.Fail("Invalid Index")
	}

	return player, index, success
}