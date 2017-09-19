package controllers

import (
	"fmt"
	"time"
	"encoding/json"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/data"
	"bloodtales/models"
)

var (
	// Defining this location in advance for speedy date determination
	PST = time.FixedZone("PST", -8*3600)
)

type SQLevent struct {
	Data           bson.M
}

func (eventStruct *SQLevent) GetIntField(field string) (val int64){
	val = -1
	if eventStruct.Data != nil {
		if tempval, ok := eventStruct.Data[field]; ok {
			val = int64(tempval.(float64))
		}
	}
	return
}

func (eventStruct *SQLevent) GetStrField(field string) (val string){
	val = ""
	if eventStruct.Data != nil {
		if tempval, ok := eventStruct.Data[field]; ok {
			val = tempval.(string)
		}
	}
	return
}

func (eventStruct *SQLevent) GetFloatField(field string) (val float64){
	val = -1.0
	if eventStruct.Data != nil {
		if tempval, ok := eventStruct.Data[field]; ok {
			val = tempval.(float64)
		}
	}
	return
}

func handleTracking() {
	handleGameAPI("/tracking", system.TokenAuthentication, PostTracking)
}

func PostTracking(context *util.Context) {
	// parse parameters
	for i := 0; i < data.GameplayConfig.BatchLimit; i++ {
		event := context.Params.GetString(fmt.Sprintf("event%d", i), "")
		if event == "" { //no more events have been sent, so we can break out of our loop
			break 
		}

		dataJson := context.Params.GetString(fmt.Sprintf("data%d", i), "")
	
		// process data
		var data bson.M = nil
		if dataJson != "" {
			util.Must(json.Unmarshal([]byte(dataJson), &data))
		}
	
		// insert tracking
		if util.HasSQLDatabase() {
			InsertTrackingSQL(context, event, 0, "", "", 0, 0, data)
		} else {
			InsertTracking(context, event, data, 0)
		}
	}
}

func InsertTracking(context *util.Context, event string, data bson.M, expireAfterHours int) {
	// get user
	user := system.GetUser(context)

	// create tracking
	tracking := &models.Tracking {
		UserID: user.ID,
		Event: event,
		Data: data,
	}

	// expiration
	if expireAfterHours > 0 {
		tracking.ExpireTime = time.Now().Add(time.Hour * time.Duration(expireAfterHours))
	}

	// insert tracking
	util.Must(tracking.Insert(context))
	return
}

func InsertTrackingSQL(context *util.Context, event string, timeId int64, itemId string, description string, count int, amount float64, data bson.M) {
	if util.HasSQLDatabase() {
		sqlEvent := SQLevent{Data: data}
		user := system.GetUser(context)
		userId := user.ID.Hex()
		timeNs := sqlEvent.GetIntField("eventTime")
		var eventTime time.Time
		if timeNs > 0 {
			eventTime = time.Unix(0,timeNs)
		} else {
			eventTime = time.Now()
		}
		//location, _ := time.LoadLocation("PST8PDT")
		dateStr := eventTime.In(PST).Format("2006-01-02")
		timeStr := eventTime.Format("2006-01-02 15:04:05z")
		factInsertStr := "insert into facts values ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

		var err error
		switch event {
		case "navigation":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, sqlEvent.GetIntField("sessionId"), sqlEvent.GetStrField("pageName"), 
				sqlEvent.GetStrField("action"),  sqlEvent.GetIntField("count"), sqlEvent.GetFloatField("duration"))
			util.Must(err)
		case "tabChange":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, sqlEvent.GetIntField("sessionId"), sqlEvent.GetStrField("pageName"), 
				sqlEvent.GetStrField("action"),  sqlEvent.GetIntField("count"), sqlEvent.GetFloatField("duration"))
			util.Must(err)
		case "infoPopup":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, sqlEvent.GetIntField("sessionId"), sqlEvent.GetStrField("category"), 
				sqlEvent.GetStrField("details"),  sqlEvent.GetIntField("count"), sqlEvent.GetFloatField("duration"))
			util.Must(err)
		case "tutorialPageExit":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, sqlEvent.GetIntField("sessionId"), itemId, 
				sqlEvent.GetStrField("description"),  sqlEvent.GetIntField("pageNumber"), sqlEvent.GetFloatField("duration"))
			util.Must(err)
		case "playerBattleSummary":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, 
				sqlEvent.GetIntField("time"), sqlEvent.GetStrField("leaderCardId"), sqlEvent.GetStrField("endResult"), 
				sqlEvent.GetIntField("score"), sqlEvent.GetFloatField("durationSec"))
			util.Must(err)
			
			pbsInsertStr := "insert into player_battle_summary values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)"
			_, err = context.SQL.Exec(pbsInsertStr,
				dateStr, timeStr, event, userId, sqlEvent.GetIntField("time"), 
				sqlEvent.GetStrField("userBattleId"), sqlEvent.GetStrField("leaderCardId"), sqlEvent.GetIntField("userLevel"), 
				sqlEvent.GetIntField("userRank"), sqlEvent.GetStrField("gameMode"), sqlEvent.GetStrField("userTeam"), 
				sqlEvent.GetFloatField("durationSec"), sqlEvent.GetIntField("score"), sqlEvent.GetStrField("endResult"), 
				sqlEvent.GetStrField("endPhase"))
			util.Must(err)
		case "cardBattleSummary":
			// This is the only one that doesn't also go into the fact table
			cbsInsertStr := "insert into card_battle_summary values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)"
			_, err = context.SQL.Exec(cbsInsertStr,
				dateStr, timeStr, event, userId, sqlEvent.GetIntField("time"), 
				sqlEvent.GetStrField("userBattleId"), sqlEvent.GetStrField("cardId"), sqlEvent.GetIntField("level"), 
				sqlEvent.GetIntField("manaCost"), sqlEvent.GetFloatField("damageManaValue"), sqlEvent.GetIntField("numUses"), 
				sqlEvent.GetFloatField("characterDamage"), sqlEvent.GetIntField("characterKills"), sqlEvent.GetFloatField("towerDamage"), 
				sqlEvent.GetIntField("towerKills"), sqlEvent.GetFloatField("totalUnitTime"), sqlEvent.GetIntField("totalUnits") )
			util.Must(err)
		case "tutorialCompleted":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, sqlEvent.GetIntField("sessionId"), itemId, 
				sqlEvent.GetStrField("description"),  count, amount)
			util.Must(err)
		case "applicationPaused":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, sqlEvent.GetIntField("sessionId"), sqlEvent.GetStrField("pageName"), 
				sqlEvent.GetStrField("description"),  sqlEvent.GetIntField("pageNumber"), sqlEvent.GetFloatField("duration"))
			util.Must(err)
		case "purchase":
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, timeId, itemId, description, count, amount)
			util.Must(err)
			
			purchInsertStr := "insert into purchase values ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
			_, err = context.SQL.Exec(purchInsertStr,
				dateStr, timeStr, event, userId, timeId, 
				sqlEvent.GetStrField("productId"), sqlEvent.GetStrField("currency"), sqlEvent.GetStrField("receipt"), 
				sqlEvent.GetFloatField("price") )
		default:
			_, err = context.SQL.Exec(factInsertStr, dateStr, timeStr, event, userId, timeId, itemId, description, count, amount)
			util.Must(err)
		}
	}
	return
}

func TrackRewards(context *util.Context, reward *models.Reward) {
	currentTime := util.TimeToTicks(time.Now().UTC())

	for i, id := range reward.Cards {
		InsertTracking(context, "gainItem", bson.M { "time":currentTime,
													 "itemId":data.ToDataName(id),
													 "type":"Card",
													 "count":reward.NumRewarded[i] }, 0)
	}

	if reward.StandardCurrency > 0 || reward.OverflowCurrency > 0 {
		InsertTracking(context, "gainItem", bson.M { "time":currentTime,
													 "itemId":"",
													 "type":"Standard",
													 "count":reward.StandardCurrency + reward.OverflowCurrency }, 0)
	}

	if reward.PremiumCurrency > 0 {
		InsertTracking(context, "gainItem", bson.M { "time":currentTime,
													 "itemId":"",
													 "type":"Premium",
													 "count":reward.PremiumCurrency }, 0)
	}
}

func TrackRewardsSQL(context *util.Context, reward *models.Reward, timeId int64) {
	currentTime := timeId
	if (currentTime <= 0){
		currentTime = util.TimeToTicks(time.Now().UTC())
	}

	for i, id := range reward.Cards {
		InsertTrackingSQL(context, "gainItem", currentTime, data.ToDataName(id), "Card", reward.NumRewarded[i], 0, nil)
	}

	if reward.StandardCurrency > 0 || reward.OverflowCurrency > 0 {
		InsertTrackingSQL(context, "gainItem", currentTime, "", "Standard", 1, float64(reward.StandardCurrency + reward.OverflowCurrency), nil)
	}

	if reward.PremiumCurrency > 0 {
		InsertTrackingSQL(context, "gainItem", currentTime, "", "Premium", 1, float64(reward.PremiumCurrency), nil)
	}
}