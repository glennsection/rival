package controllers

import (
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
	notifications, err := models.GetReceivedNotifications(context.DB, player, types)
	util.Must(err)

	// insert all sender names
	for _, notification := range notifications {
		util.Must(prepareNotification(context, notification))
	}

	// result
	context.SetData("notifications", notifications)
}

func RespondNotification(context *util.Context) {
	// parse parameters
	notificationID := context.Params.GetRequiredId("id")
	action := context.Params.GetRequiredString("action")

	// get notification
	notification, err := models.GetNotificationById(context.DB, notificationID)
	util.Must(err)

	util.Must(respondNotification(context, notification, action))
}

func ViewNotifications(context *util.Context) {
	// parse parameters
	notificationIDs := context.Params.GetRequiredIds("ids")

	// handle notifications as viewed
	models.ViewNotificationsByIds(context.DB, notificationIDs)
}

// TODO - migrate these into a more generic/universal system...
func prepareNotification(context *util.Context, notification *models.Notification) (err error) {
	// get sender name
	if notification.SenderID.Valid() {
		notification.SenderName = GetPlayerName(context, notification.SenderID)
	} else {
		notification.SenderName = "" // internal message?
	}

	switch notification.Type {

	case "FriendRequest":
		// TODO - add PlayerClient into data

	}

	return
}

func respondNotification(context *util.Context, notification *models.Notification, action string) (err error) {
	switch notification.Type {

	case "FriendRequest":
		// handle friend request
		if action == "accept" {
			// get sender friends
			var senderFriends *models.Friends
			senderFriends, err = models.GetFriendsByPlayerId(context.DB, notification.SenderID, true)
			if err != nil {
				return
			}
			// append sender friends
			senderFriends.FriendIDs = append(senderFriends.FriendIDs, notification.ReceiverID)
			err = senderFriends.Save(context.DB)
			if err != nil {
				return
			}

			// get receiver friends
			var receiverFriends *models.Friends
			receiverFriends, err = models.GetFriendsByPlayerId(context.DB, notification.ReceiverID, true)
			if err != nil {
				return
			}
			// append receiver friends
			receiverFriends.FriendIDs = append(receiverFriends.FriendIDs, notification.SenderID)
			err = receiverFriends.Save(context.DB)
			if err != nil {
				return
			}

			// notify sender
			context.SocketSend("FriendAccepted")
		}

	}

	err = notification.Delete(context.DB)

	return
}