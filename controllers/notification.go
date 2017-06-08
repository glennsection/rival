package controllers

import (
	"fmt"
	"strings"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
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

	// insert all sender names
	for _, notification := range notifications {
		prepareNotification(context, notification)
	}

	// result
	context.SetData("notifications", notifications)
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

	case "FriendBattle":
		// handle friend battle
		respondFriendBattle(context, notification, action)

	}

	// notify sender
	if notification.SenderID.Valid() {
		senderUserID := models.GetUserIdByPlayerId(context, notification.SenderID)
		system.SocketSend(senderUserID, fmt.Sprintf("%s-%s", notification.Type, action), nil)
	}

	// delete notification
	util.Must(notification.Delete(context))
}