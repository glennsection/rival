package controllers 

import(
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/data"
	"bloodtales/models"
)

func handleDebug() {
	handleGameAPI("/debug/addTome", system.TokenAuthentication, DebugAddTome)
	handleGameAPI("/debug/addVictoryTome", system.TokenAuthentication, DebugAddNextVictoryTome)
	handleGameAPI("/debug/addCards", system.TokenAuthentication, DebugAddCards)
	handleGameAPI("/debug/addPremiumCurrency", system.TokenAuthentication, DebugAddPremiumCurrency)
	handleGameAPI("/debug/addStandardCurrency", system.TokenAuthentication, DebugAddStandardCurrency)
	handleGameAPI("/debug/setRank", system.TokenAuthentication, DebugSetRank)
	handleGameAPI("/debug/refreshStore", system.TokenAuthentication, DebugRefreshStore)
	handleGameAPI("/debug/clearStoreHistory", system.TokenAuthentication, DebugClearStoreHistory)
}

func DebugAddTome(context *util.Context) {
	tomeId := context.Params.GetRequiredString("tomeId")
	tomeDataId := data.ToDataId(tomeId)

	if tomeData := data.GetTome(tomeDataId); tomeData == nil {
		context.Fail("Invalid tome ID")
		return
	}

	player := GetPlayer(context)
	_, tome := player.GetEmptyTomeSlot()
	if tome == nil {
		context.Fail("No available tome slots")
		return
	}

	tome.DataID = tomeDataId
	tome.State = models.TomeLocked
	tome.UnlockTime = 0
	tome.League = data.GetLeague(data.GetRank(player.RankPoints).Level)

	player.SetDirty(models.PlayerDataMask_Tomes)
	player.Save(context)
}

func DebugAddNextVictoryTome(context *util.Context) {
	winCount := context.Params.GetRequiredInt("winCount")
	if winCount < 0 {
		context.Fail("Invalid Request")
	}

	player := GetPlayer(context)
	_, tome := player.GetEmptyTomeSlot()
	if tome == nil {
		context.Fail("No available tome slots")
		return
	}

	tome.DataID = data.GetNextVictoryTomeID(winCount)
	tome.State = models.TomeLocked
	tome.UnlockTime = 0

	player.SetDirty(models.PlayerDataMask_Tomes)
	player.Save(context)
}

func DebugAddCards(context *util.Context) {
	cardId := context.Params.GetRequiredString("cardId")
	count := context.Params.GetRequiredInt("count")

	cardDataId := data.ToDataId(cardId)
	if !isValidCardId(cardDataId) {
		context.Fail("Invalid Id")
		return
	}

	if count <= 0 {
		context.Fail("Invalid card count\nCount must be non-zero and positive")
		return
	}

	player := GetPlayer(context)
	player.AddCards(cardDataId, count)

	player.SetDirty(models.PlayerDataMask_Cards)
	player.Save(context)
}

func isValidCardId(cardDataId data.DataId) bool {
	cardIds := data.GetCards(func(card *data.CardData) bool { return true })

	for _, id := range cardIds {
		if id == cardDataId {
			return true
		}
	}

	return false
}

func DebugAddPremiumCurrency(context *util.Context) {
	amount := context.Params.GetRequiredInt("amount")

	if amount <= 0 {
		context.Fail("Invalid amount\nAmount must be non-zero and positive")
		return
	}

	player := GetPlayer(context)

	player.PremiumCurrency += amount

	player.Save(context)
	player.SetDirty(models.PlayerDataMask_Currency)
}

func DebugAddStandardCurrency(context *util.Context) {
	amount := context.Params.GetRequiredInt("amount")

	if amount <= 0 {
		context.Fail("Invalid amount\nAmount must be non-zero and positive")
		return
	}

	player := GetPlayer(context)
	
	player.StandardCurrency += amount

	player.Save(context)
	player.SetDirty(models.PlayerDataMask_Currency)
}

func DebugSetRank(context *util.Context) {
	stars := context.Params.GetRequiredInt("stars")

	if stars <= 0 {
		context.Fail("Invalid amount\nAmount must be non-zero and positive")
		return
	}

	player := GetPlayer(context)

	player.RankPoints = stars

	player.Save(context)
	player.SetDirty(models.PlayerDataMask_Stars)
}

func DebugRefreshStore(context *util.Context) {
	player := GetPlayer(context)

	player.Store.SpecialOffer.ExpirationDate = 0

	for i, _ := range player.Store.Cards {
		player.Store.Cards[i].ExpirationDate = 0
	}

	player.Save(context)
}

func DebugClearStoreHistory(context *util.Context) {
	player := GetPlayer(context)

	player.Store.OneTimePurchaseHistory = map[string]models.OfferHistory {}
	DebugRefreshStore(context)
}