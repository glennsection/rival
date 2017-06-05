package controllers

import (
	"bloodtales/system"
	"bloodtales/util"
)

func HandleGame() {
	handleUser()
	handlePlayer()
	handleTome()
	handleCard()
	handleDeck()
	handleMatch()
	handlePurchase()
	handleNotification()
	handleStore()
	handleFriends()
	handleGuild()
	handleTracking()
}

func handleGameAPI(pattern string, authType system.AuthenticationType, handler func(*util.Context)) {
	// all template requests start here
	system.App.HandleAPI(pattern, authType, func(context *util.Context) {
		handler(context)

		// handle player data deltas
		player := GetPlayer(context)
		if player != nil {
			playerData := player.MarshalDirty(context)

			if playerData != nil {
				context.SetData("playerData", playerData)
				context.SetData("playerDataMask", player.DirtyMask)
			}
		}
	})
}
