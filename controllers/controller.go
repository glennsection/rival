package controllers

import (
	"bloodtales/system"
	"bloodtales/models"
)

func HandleGameAPI(application *system.Application, pattern string, authType system.AuthenticationType, handler func(*system.Context)) {
	// all template requests start here
	application.HandleAPI(pattern, authType, func(context *system.Context) {
		handler(context)

		// handle dirty flags in context.UpdatedData
		if context.UpdateMask != models.UpdateMask_None {
			handleUpdateMask(context)
		}
	})
}

func handleUpdateMask(context *system.Context) {
	if context.PlayerData == nil {
		context.PlayerData = map[string]interface{} {}
	}

	player := GetPlayer(context);
	player.HandleUpdateMask(context.UpdateMask, &context.PlayerData)
}
