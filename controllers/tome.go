package controllers

import (
	"strconv"
	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/system"
)

func HandleTome(application *system.Application) {
	application.HandleAPI("/tome/unlock", system.TokenAuthentication, UnlockTome)
	application.HandleAPI("/tome/open", system.TokenAuthentication, OpenTome)
	application.HandleAPI("/tome/rush", system.TokenAuthentication, RushTome)
}

func UnlockTome(context *system.Context) {
	//Validate the request
	index, player, valid := ValidateTomeRequest(context)
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

		err := player.Update(context.DB)
		if err != nil {
			panic(err)
			return
		}

		var data *models.Tome 
		data = &player.Tomes[index]
		context.Data = data
	} else {
		context.Fail("Already unlocking a tome.")
	}
}

func OpenTome(context *system.Context) {
	//Validate the request
	index, player, valid := ValidateTomeRequest(context)
	if !valid {
		return
	}

	// check to see if the tome is ready to open
	(&player.Tomes[index]).UpdateTome()
	if player.Tomes[index].State != models.TomeUnlocked {
		context.Fail("Tome not ready")
		return
	}

	// TODO add cards recieved from tome to context data
	(&player.Tomes[index]).OpenTome()

	err := player.Update(context.DB)
	if err != nil {
		panic(err)
		return
	}

	context.Data = &player.Tomes[index]
}

func RushTome(context *system.Context) {
	//Validate the request
	index, player, valid := ValidateTomeRequest(context)
	if !valid {
		return
	}

	cost := data.GetTome(player.Tomes[index].DataID).GemsToUnlock

	// check to see if the player has enough premium currency
	if cost > player.PremiumCurrency {
		context.Fail("Not enough premium currency")
		return
	}

	player.PremiumCurrency -= cost
	(player.Tomes[index]).OpenTome()

	err := player.Update(context.DB)
	if err != nil {
		panic(err)
		return
	}

	context.Data = &player.Tomes[index]
}

func ValidateTomeRequest(context *system.Context) (index int, player *models.Player, success bool) {
	// initialize values
	index = -1
	player = context.GetPlayer()
	success = false

	// parse parameters
	tomeId := context.GetRequiredParameter("tomeId")
	
	index, err := strconv.Atoi(tomeId)
	if err != nil {
		panic(err)
		return 
	}

	// make sure the tome exists
	if index >= len(player.Tomes) || index < 0 || player.Tomes[index].State == models.TomeEmpty {
		context.Fail("Tome does not exist")
		return 
	}

	success = true
	return
}