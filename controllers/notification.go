package controllers

import (
	"fmt"
	"time"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handleNotification() {
	handleGameAPI("/notification/get", system.TokenAuthentication, GetNotifications)
	handleGameAPI("/notification/respond", system.TokenAuthentication, RespondNotification)
	handleGameAPI("/notification/view", system.TokenAuthentication, ViewNotifications)

	handleGameAPI("/notification/test", system.TokenAuthentication, TestNotifications)
}

func GetNotifications(context *util.Context) {
	// current user and player
	user := system.GetUser(context)
	player := GetPlayer(context)

	// get all notifications for user
	notifications, err := models.GetReceivedNotifications(context.DB, user, player)
	util.Must(err)

	// insert all sender names
	for _, notification := range notifications {
		// get sender name
		if notification.SenderID.Valid() {
			notification.SenderName = GetUserName(context, notification.SenderID)
		} else {
			notification.SenderName = "" // internal message?
		}
	}

	// result
	context.SetData("notification", struct { Notifications []*models.Notification } { Notifications: notifications, })
}

func RespondNotification(context *util.Context) {
	// parse parameters
	// notificationID := context.Params.GetRequiredId("id")
	// action := context.Params.GetRequiredString("action")

	// TODO
}

func ViewNotifications(context *util.Context) {
	// parse parameters
	notificationIDs := context.Params.GetRequiredIds("ids")

	// handle notifications as viewed
	models.ViewNotificationsByIds(context.DB, notificationIDs)
}

func TestNotifications(context *util.Context) {
	// current user and player
	user := system.GetUser(context)
	player := GetPlayer(context)

	// create test notifications
	util.Must((&models.Notification {
		//SenderID: nil,
		ReceiverID: user.ID,
		Guild: false,
		ExpiresAt: time.Now().Add(time.Minute),
		//Type: "",
		//Image: "",
		Message: "System to User",
		// Actions: []models.NotificationAction {
		// 	models.NotificationAction {
		// 		Name: "Accept",
		// 		URL: "/notification/respond?action=accept",
		// 	},
		// 	models.NotificationAction {
		// 		Name: "Decline",
		// 		URL: "/notification/respond?action=decline",
		// 	},
		// },
		//Data: bson.M {},
	}).Save(context.DB))

	util.Must((&models.Notification {
		SenderID: user.ID,
		ReceiverID: user.ID,
		Guild: false,
		ExpiresAt: time.Now().Add(time.Minute),
		//Type: "",
		//Image: "",
		Message: fmt.Sprintf("User to User"),
		Actions: []models.NotificationAction {
			models.NotificationAction {
				Name: "Accept",
				URL: "/notification/respond?action=accept",
			},
			models.NotificationAction {
				Name: "Decline",
				URL: "/notification/respond?action=decline",
			},
		},
		//Data: bson.M {},
	}).Save(context.DB))

	util.Must((&models.Notification {
		SenderID: user.ID,
		//ReceiverID: nil,
		Guild: false,
		ExpiresAt: time.Now().Add(time.Minute),
		//Type: "",
		//Image: "",
		Message: "User to All",
		// Actions: []models.NotificationAction {
		// 	models.NotificationAction {
		// 		Name: "Accept",
		// 		URL: "/notification/respond?action=accept",
		// 	},
		// 	models.NotificationAction {
		// 		Name: "Decline",
		// 		URL: "/notification/respond?action=decline",
		// 	},
		// },
		//Data: bson.M {},
	}).Save(context.DB))

	util.Must((&models.Notification {
		SenderID: user.ID,
		ReceiverID: player.GuildID,
		Guild: true,
		ExpiresAt: time.Now().Add(time.Minute),
		//Type: "",
		//Image: "",
		Message: "User to Guild",
		Actions: []models.NotificationAction {
			models.NotificationAction {
				Name: "Accept",
				URL: "/notification/respond?action=accept",
			},
		},
		//Data: bson.M {},
	}).Save(context.DB))
}
