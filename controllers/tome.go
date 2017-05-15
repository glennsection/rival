package controllers

import (
	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
)

func HandleTome(application *system.Application) {
	HandleGameAPI(application, "/tome/unlock", system.TokenAuthentication, UnlockTome)
	HandleGameAPI(application, "/tome/open", system.TokenAuthentication, OpenTome)
	HandleGameAPI(application, "/tome/rush", system.TokenAuthentication, RushTome)
	HandleGameAPI(application, "/tome/free", system.TokenAuthentication, ClaimFreeTome)
	HandleGameAPI(application, "/tome/arena", system.TokenAuthentication, ClaimArenaTome)
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

		util.Must(player.Save(context.DB))

		context.SetDirty([]int64{models.UpdateMask_Tomes})
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

	reward, err := player.AddRewards(context.DB, &player.Tomes[index]) 
	util.Must(err)

	context.SetDirty([]int64{models.UpdateMask_Currency,
							 models.UpdateMask_Cards, 
							 models.UpdateMask_Tomes})
	context.Data = reward
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

	reward, err := player.AddRewards(context.DB, &player.Tomes[index]) 
	util.Must(err)

	context.SetDirty([]int64{models.UpdateMask_Currency,
							 models.UpdateMask_Cards, 
							 models.UpdateMask_Tomes})
	context.Data = reward
}

func ClaimFreeTome(context *system.Context) {
	player := GetPlayer(context)
	reward, err := player.ClaimFreeTome(context.DB)
	util.Must(err)

	if reward == nil {
		context.Fail("No free tomes available")
		return
	}

	context.SetDirty([]int64{models.UpdateMask_Currency,
							 models.UpdateMask_Cards, 
							 models.UpdateMask_Tomes})
	context.Data = reward
}

func ClaimArenaTome(context *system.Context) {
	player := GetPlayer(context)
	reward, err := player.ClaimArenaTome(context.DB)
	util.Must(err)

	if reward == nil {
		context.Fail("Not enough arena points")
		return
	}

	context.SetDirty([]int64{models.UpdateMask_Currency,
							 models.UpdateMask_Cards, 
							 models.UpdateMask_Tomes})
	context.Data = reward
}

func ValidateTomeRequest(context *system.Context) (index int, player *models.Player, success bool) {
	// initialize values
	player = GetPlayer(context)
	success = false

	// parse parameters
	index = context.Params.GetRequiredInt("tomeId")

	// make sure the tome exists
	if index >= len(player.Tomes) || index < 0 || player.Tomes[index].State == models.TomeEmpty {
		context.Fail("Tome does not exist")
		return 
	}

	success = true
	return
}