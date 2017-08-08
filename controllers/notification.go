package controllers

import (
	"fmt"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
)

func handleNotification() {
	handleGameAPI("/notifications/get", system.TokenAuthentication, GetNotifications)
	handleGameAPI("/notifications/respond", system.TokenAuthentication, RespondNotification)
	handleGameAPI("/notifications/view", system.TokenAuthentication, ViewNotifications)
}

func GetNotifications(context *util.Context) {
	// parse parameters
	typesParam := context.Params.GetString("types", "")

	// current player
	player := GetPlayer(context)

	// parse types filter
	var types []string
	if typesParam != "" {
		types = strings.Split(typesParam, ",")
	}

	// get all notifications for player
	notifications, err := models.GetReceivedNotifications(context, player, types)
	util.Must(err)

	var senderPlayerIds []bson.ObjectId
	var senderPlayers []*models.Player
	var playerClientList []*models.PlayerClient

	util.Must(err)

	// insert all sender names
	for _, notification := range notifications {
		prepareNotification(context, notification)
		senderPlayerIds = append(senderPlayerIds, notification.SenderID)
		//var senderPlayer *models.Player
		//err = context.DB.C(models.PlayerCollectionName).Find(bson.M{"_id": notification.SenderID}).One(&senderPlayer)
		//util.Must(err)
	}

	err = context.DB.C(models.PlayerCollectionName).Find(bson.M{"_id": bson.M{"$in": senderPlayerIds}}).All(&senderPlayers)
	util.Must(err)

	for _, senderPlayer := range senderPlayers {
		playerClient, err := senderPlayer.GetPlayerClient(context)
		util.Must(err)

		playerClientList = append(playerClientList, playerClient)
	}

	// result
	context.SetData("notifications", notifications)
	context.SetData("senderClientData", playerClientList)
}

func RespondNotification(context *util.Context) {
	// parse parameters
	notificationID := context.Params.GetRequiredId("id")
	action := context.Params.GetRequiredString("action")

	// get notification
	notification, err := models.GetNotificationById(context, notificationID)
	util.Must(err)

	respondNotification(context, notification, action)
}

func ViewNotifications(context *util.Context) {
	// parse parameters
	notificationIDs := context.Params.GetRequiredIds("ids")

	// handle notifications as viewed
	models.ViewNotificationsByIds(context, notificationIDs)
}

// TODO - migrate these into a more generic/universal system...
func prepareNotification(context *util.Context, notification *models.Notification) {
	// get sender name
	if notification.SenderID.Valid() {
		notification.SenderName = models.GetPlayerName(context, notification.SenderID)
	} else {
		notification.SenderName = "" // "System Message"?
	}

	switch notification.Type {

	case "FriendRequest":
	case "FriendBattle":
		prepareFriendNotification(context, notification)

	}
}

func respondNotification(context *util.Context, notification *models.Notification, action string) {
	switch notification.Type {

	case "FriendRequest":
		// handle friend request
		respondFriendRequest(context, notification, action)

		// notify sender
		if notification.SenderID.Valid() {
			senderUserID := models.GetUserIdByPlayerId(context, notification.SenderID)
			system.SocketSend(senderUserID, fmt.Sprintf("%s-%s", notification.Type, action), nil)
		}


	case "FriendBattle":
		// handle friend battle
		respondFriendBattle(context, notification, action)
		
		// notify sender
		if notification.SenderID.Valid() {
			senderUserID := models.GetUserIdByPlayerId(context, notification.SenderID)
			system.SocketSend(senderUserID, fmt.Sprintf("%s-%s", notification.Type, action), nil)
		}


	case "GuildBattle":
		// handle Guild Battle
		respondGuildBattle(context, notification, action)

		if notification.SenderID.Valid() {
			guild, err := models.GetGuildById(context, GetPlayer(context).GuildID)
			util.Must(err)

			var memberPlayers []*models.Player
			err = context.DB.C(models.PlayerCollectionName).Find(bson.M{"gd": guild.ID}).All(&memberPlayers)
			util.Must(err)

			for _, memberPlayer := range memberPlayers {
				// notify receiver
				socketData := map[string]interface{}{"notificationId": notification.ID}
				system.SocketSend(memberPlayer.UserID, "GuildBattle-clear", socketData)
			}
		}

		break
	}
	
	// delete notification
	util.Must(notification.Delete(context))
}
