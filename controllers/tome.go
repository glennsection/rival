package controllers

import (
	"gopkg.in/mgo.v2/bson"

	"bloodtales/data"
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
)

func handleTome() {
	handleGameAPI("/tome/unlock", system.TokenAuthentication, UnlockTome)
	handleGameAPI("/tome/open", system.TokenAuthentication, OpenTome)
	handleGameAPI("/tome/rush", system.TokenAuthentication, RushTome)
	handleGameAPI("/tome/free", system.TokenAuthentication, ClaimFreeTome)
	handleGameAPI("/tome/arena", system.TokenAuthentication, ClaimArenaTome)
}

func UnlockTome(context *util.Context) {
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

		util.Must(player.Save(context))

		player.SetDirty(models.PlayerDataMask_Tomes)
	} else {
		context.Fail("Already unlocking a tome.")
	}
}

func OpenTome(context *util.Context) {
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

	// analytics
	InsertTracking(context, "tomeOpened", bson.M { "rarity": player.Tomes[index].GetData().Rarity }, 0)

	reward, err := player.AddTomeRewards(context, &player.Tomes[index]) 
	util.Must(err)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
	context.SetData("reward", reward)

	TrackRewards(context, reward)
}

func RushTome(context *util.Context) {
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

	// analytics
	InsertTracking(context, "tomeOpened", bson.M { "rarity": player.Tomes[index].GetData().Rarity }, 0)

	player.PremiumCurrency -= cost

	reward, err := player.AddTomeRewards(context, &player.Tomes[index]) 
	util.Must(err)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
	context.SetData("reward", reward)

	TrackRewards(context, reward)
}

func ClaimFreeTome(context *util.Context) {
	player := GetPlayer(context)
	reward, err := player.ClaimFreeTome(context)
	util.Must(err)

	if reward == nil {
		context.Fail("No free tomes available")
		return
	}

	// analytics
	InsertTracking(context, "tomeOpened", bson.M { "rarity": "Free" }, 0)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
	context.SetData("reward", reward)

	TrackRewards(context, reward)
}

func ClaimArenaTome(context *util.Context) {
	player := GetPlayer(context)
	reward, err := player.ClaimArenaTome(context)
	util.Must(err)

	if reward == nil {
		context.Fail("Not enough arena points")
		return
	}

	// analytics
	InsertTracking(context, "tomeOpened", bson.M { "rarity": "Arena" }, 0)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes)
	context.SetData("reward", reward)

	TrackRewards(context, reward)
}

func ValidateTomeRequest(context *util.Context) (index int, player *models.Player, success bool) {
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