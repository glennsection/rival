package controllers

import (
	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handleReplay() {
	handleGameAPI("/replay/list", system.TokenAuthentication, GetReplayList)
	handleGameAPI("/replay/get", system.TokenAuthentication, GetReplay)
	handleGameAPI("/replay/getLastReplay", system.TokenAuthentication, GetLastReplayInfoByUser)
	handleGameAPI("/replay/getTopReplays", system.TokenAuthentication, GetTopReplays)
	handleGameAPI("/replay/set", system.TokenAuthentication, SetReplay)
	handleGameAPI("/replay/delete", system.TokenAuthentication, DeleteReplay)
}

func GetReplayList(context *util.Context) {
	replayInfos, err := models.GetReplayInfosByUser(context, context.UserID)
	util.Must(err)

	context.SetData("replays", replayInfos)
}

func GetTopReplays(context *util.Context) {
	replayInfos, err := models.GetTopReplays(context, context.UserID)
	util.Must(err)

	context.SetData("replays", replayInfos)
}

func GetLastReplayInfoByUser(context *util.Context) {
	playerTag := context.Params.GetRequiredString("tag")

	player,err := models.GetPlayerByTag(context, playerTag)
	util.Must(err)

	replayInfo, err2 := models.GetLastReplayInfoByUser(context, player.UserID)
	util.Must(err2)

	context.SetData("replay", replayInfo)
}

func GetReplay(context *util.Context) {
	// parse parameters
	infoId := context.Params.GetRequiredId("id")

	replayData, err := models.GetReplayDataByInfo(context, infoId)
	util.Must(err)

	context.SetData("replayData", replayData)
}

func SetReplay(context *util.Context) {
	// parse parameters
	info := context.Params.GetRequiredString("info")
	data := context.Params.GetRequiredString("data")
	rank := context.Params.GetRequiredInt("rank")

	util.Must(models.CreateReplay(context, info, data, rank))
}

func DeleteReplay(context *util.Context) {
	infoId := context.Params.GetRequiredId("id")

	replayInfo, err := models.GetReplayInfoById(context, infoId)
	util.Must(err)

	replayInfo.Delete(context)
}
