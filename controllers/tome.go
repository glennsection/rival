package controllers

import (
	"bloodtales/system"
)

func HandleTome(application *system.Application) {
	//application.Handle("/tome/unlock", system.TokenAuthentication, UnlockTome)
}

func UnlockTome(session *system.Session) {
	// parse parameters
	//tomeId := session.GetRequiredParameter("tomeId")

	// get player
	//player := session.GetPlayer()

	// TODO - start unlock...
}
