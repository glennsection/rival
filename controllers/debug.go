package controllers 

import(
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/data"
	"bloodtales/models"
)

func handleDebug() {
	handleGameAPI("/debug/addTome", system.TokenAuthentication, DebugAddTome)
	handleGameAPI("/debug/addVictoryTome", system.TokenAuthentication, DebugAddNextVictoryTome)
}

func DebugAddTome(context *util.Context) {
	tomeId := context.Params.GetRequiredString("tomeId")
	tomeDataId := data.ToDataId(tomeId)

	if tomeData := data.GetTome(tomeDataId); tomeData == nil {
		context.Fail("Invalid tome ID")
		return
	}

	player := GetPlayer(context)
	_, tome := player.GetEmptyTomeSlot()
	if tome == nil {
		context.Fail("No available tome slots")
		return
	}

	tome.DataID = tomeDataId
	tome.State = models.TomeLocked
	tome.UnlockTime = 0

	player.SetDirty(models.PlayerDataMask_Tomes)
	player.Save(context)
}

func DebugAddNextVictoryTome(context *util.Context) {
	winCount := context.Params.GetRequiredInt("winCount")
	if winCount < 0 {
		context.Fail("Invalid Request")
	}

	player := GetPlayer(context)
	_, tome := player.GetEmptyTomeSlot()
	if tome == nil {
		context.Fail("No available tome slots")
		return
	}

	tome.DataID = data.GetNextVictoryTomeID(winCount)
	tome.State = models.TomeLocked
	tome.UnlockTime = 0

	player.SetDirty(models.PlayerDataMask_Tomes)
	player.Save(context)
}