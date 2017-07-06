package controllers

import (
	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handleReplay() {
	handleGameAPI("/replay/list", system.TokenAuthentication, GetReplayList)
	handleGameAPI("/replay/get", system.TokenAuthentication, GetReplay)
	handleGameAPI("/replay/set", system.TokenAuthentication, SetReplay)
}

func GetReplayList(context *util.Context) {
	replayInfos, err := models.GetReplayInfosByUser(context, context.UserID)
	util.Must(err)

	context.SetData("replays", replayInfos)
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

	util.Must(models.CreateReplay(context, info, data))
}
