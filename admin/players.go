package admin

import (
	"bloodtales/system"
	"bloodtales/models"
)

func ShowPlayers(session *system.Session) {
	// parse parameters
	page := session.GetIntParameter("page", 1)

	// get paginated players
	var players []models.Player
	err := models.Paginate(session.Application.DB.C(models.PlayerCollectionName).Find(nil), DefaultPageSize, page).All(&players)
	if err != nil {
		panic(err)
	}

	session.Data = players
}

func ShowPlayer(session *system.Session) {
	// parse parameters
	// playerId := session.GetRequiredParameter("playerId")

	// session.Data = models.GetPlayerById(session.Application.DB, playerId)
}
