package controllers

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/models"
)

func SetPlayer(session *system.Session) {
	// parse parameters
	//data := session.GetRequiredParameter("data")
	// HACK !!!!!!!!!!!!!!!!!!!
	data := "{\"standardCurrency\"=100}"

	// parse data
	player, err := models.ParsePlayer(data)
	if err != nil {
		panic(err)
	}

	// bind to user
	player.UserID = session.User.ID

	// set data
	if err = models.SetPlayer(session.Application.DB, player); err != nil {
		panic(err)
	}

	session.Message("Player set successfully")
}

func GetPlayer(session *system.Session) {
	// get player
	player, _ := models.GetPlayerByUser(session.Application.DB, session.User.ID)
	if player != nil {
		session.Message("Found player")
	} else {
		panic(fmt.Sprintf("Failed to find player for username: %v", session.User.Username))
	}
}
