package controllers

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
)

func SetPlayer(session *system.Session) {
	// parse parameters
	data := session.GetRequiredParameter("data")

	// update data
	if err := models.UpdatePlayer(session.Application.DB, session.User.ID, data); err != nil {
		panic(err)
	}

	session.Message("Player updated successfully")
}

func GetPlayer(session *system.Session) {
	// get player
	player, _ := models.GetPlayerByUser(session.Application.DB, session.User.ID)
	if player != nil {
		// set successful response
		session.Message("Found player")
		session.Data = player
	} else {
		panic(fmt.Sprintf("Failed to find player for username: %v", session.User.Username))
	}
}
