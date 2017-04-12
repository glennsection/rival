package controllers

import (
	"strconv"
	"bloodtales/models"
	"bloodtales/system"
)

func UnlockTome(session *system.Session) {
	//Validate the request
	index, player, valid := ValidateTomeRequest(session)
	if !valid {
		return
	}

	// check to see if any tomes are unlocking
	busy := false;
	for i := 0; i < len(player.Tomes); i++ {
		if player.Tomes[i].State == models.TomeUnlocking {
			busy = true
		}
	}

	// start unlock if no other tomes are unlocking
	if !busy {
		(&player.Tomes[index]).StartUnlocking()

		err := player.Update(session.Application.DB)
		if err != nil {
			panic(err)
			return
		}

		var data *models.Tome 
		data = &player.Tomes[index]
		session.Data = data
	} else {
		session.Fail("Already unlocking a tome.")
	}
}

func OpenTome(session *system.Session) {
	//Validate the request
	index, player, valid := ValidateTomeRequest(session)
	if !valid {
		return
	}

	// check to see if the tome is ready to open
	(&player.Tomes[index]).UpdateTome()
	if player.Tomes[index].State != models.TomeUnlocked {
		session.Fail("Tome not ready")
		return
	}

	// TODO add cards recieved from tome to session data
	(&player.Tomes[index]).OpenTome()

	err := player.Update(session.Application.DB)
	if err != nil {
		panic(err)
		return
	}

	session.Data = player
}

func ValidateTomeRequest(session *system.Session) (index int, player *models.Player, success bool) {
	// initialize values
	index = -1
	player = session.GetPlayer()
	success = false

	// parse parameters
	tomeId := session.GetRequiredParameter("tomeId")
	
	index, err := strconv.Atoi(tomeId)
	if err != nil {
		panic(err)
		return 
	}

	// make sure the tome exists
	if index >= len(player.Tomes) || index < 0 || player.Tomes[index].State == models.TomeEmpty {
		session.Fail("Tome does not exist")
		return 
	}

	success = true
	return
}