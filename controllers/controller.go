package controllers

import (
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func HandleGame() {
	HandleUser()
	HandlePlayer()
	HandleTome()
	HandleCard()
	HandleDeck()
	HandleMatch()
	HandlePurchase()
	HandleNotification()
	HandleStore()
}

func HandleGameAPI(pattern string, authType system.AuthenticationType, handler func(*util.Context)) {
	// all template requests start here
	system.App.HandleAPI(pattern, authType, func(context *util.Context) {
		handler(context)

		// handle dirty flags in context.UpdatedData
		if context.UpdateMask != models.UpdateMask_None {
			handleUpdateMask(context)
		}
	})
}

func handleUpdateMask(context *util.Context) {
	if context.PlayerData == nil {
		context.PlayerData = map[string]interface{} {}
	}

	player := GetPlayer(context);
	player.HandleUpdateMask(context.UpdateMask, &context.PlayerData)
}