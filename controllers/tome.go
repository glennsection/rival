package controllers

import (
	"time"
	"strconv"
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
	handleGameAPI("/tome/guild", system.TokenAuthentication, ClaimGuildTome)
}

func UnlockTome(context *util.Context) {
	// parse parameters
	index := context.Params.GetRequiredInt("tomeId")

	// initialize values
	player := GetPlayer(context)

	// make sure the tome exists
	if index >= len(player.Tomes) || index < 0 || player.Tomes[index].State == models.TomeEmpty {
		context.Fail("Tome does not exist")
		return
	}

	// check to see if a tome is unlocking
	if player.ActiveTome.State == models.TomeUnlocking {
		context.Fail("Already unlocking a tome.")
	}

	// start unlock
	player.StartUnlocking(index)

	util.Must(player.Save(context))

	player.SetDirty(models.PlayerDataMask_Tomes)
}

func OpenTome(context *util.Context) {
	player := GetPlayer(context)

	// check to see if the tome is ready to open
	(&player.ActiveTome).UpdateTome()
	if player.ActiveTome.State != models.TomeUnlocked {
		context.Fail("Tome not ready")
		return
	}

	reward, err := player.AddTomeRewards(context, &player.ActiveTome)
	util.Must(err)

	trackTome(context, reward, data.ToDataName(player.ActiveTome.DataID), 0)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes, models.PlayerDataMask_Deck)
	context.SetData("reward", reward)
}

func RushTome(context *util.Context) {
	// parse parameters
	tomeId := context.Params.GetRequiredString("tomeId")

	// initialize values
	player := GetPlayer(context)
	var tome *models.Tome

	if tomeId == "active" {
		tome = &(player.ActiveTome)
	} else {
		var index int
		var err error

		if index, err = strconv.Atoi(tomeId); err != nil || index >= len(player.Tomes) || index < 0 || player.Tomes[index].State == models.TomeEmpty {
			context.Fail("Tome does not exist")
			return
		}

		tome = &player.Tomes[index]
	}

	cost := tome.GetUnlockCost()

	// check to see if the player has enough premium currency
	if cost > player.PremiumCurrency {
		context.Fail("Not enough premium currency")
		return
	}

	player.PremiumCurrency -= cost

	reward, err := player.AddTomeRewards(context, tome)
	util.Must(err)

	// analytics
	trackTome(context, reward, data.ToDataName(tome.DataID), cost)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes, models.PlayerDataMask_Deck)
	context.SetData("reward", reward)
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
	trackTome(context, reward, "Free", 0)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes, models.PlayerDataMask_Deck)
	context.SetData("reward", reward)
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
	trackTome(context, reward, "Arena", 0)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes, models.PlayerDataMask_Deck)
	context.SetData("reward", reward)
}

func ClaimGuildTome(context *util.Context) {
	player := GetPlayer(context)
	reward, err := player.ClaimArenaTome(context)
	util.Must(err)

	if reward == nil {
		if !player.IsInGuild() {
			context.Fail("You must be in a guild to claim this tome")
		} else {
			context.Fail("Guild tome not available")
		}
		return
	}

	// analytics
	trackTome(context, reward, "Guild", 0)

	player.SetDirty(models.PlayerDataMask_Currency, models.PlayerDataMask_Cards, models.PlayerDataMask_Tomes, models.PlayerDataMask_Deck)
	context.SetData("reward", reward)
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

func trackTome(context *util.Context, reward *models.Reward, itemId string, cost int) {
	currentTime := util.TimeToTicks(time.Now().UTC())
	if util.HasSQLDatabase() {
		InsertTrackingSQL(context, "tomeOpened", currentTime, itemId, 
			"Premium", 1, float64(cost), nil)
	} else {
		InsertTracking(context, "tomeOpened", bson.M{"tomeId": itemId,
			"gemsSpent": cost}, 0)
	}

	if util.HasSQLDatabase() {
		TrackRewardsSQL(context, reward, currentTime)
	} else {
		TrackRewards(context, reward)
	}
}