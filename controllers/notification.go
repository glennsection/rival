package controllers

import (
	"fmt"
	"time"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func HandleNotification() {
	HandleGameAPI("/notification/get", system.TokenAuthentication, GetNotifications)
	HandleGameAPI("/notification/respond", system.TokenAuthentication, RespondNotification)

	HandleGameAPI("/notification/test", system.TokenAuthentication, TestNotifications)
}

func GetNotifications(context *system.Context) {
	// current user
	user := system.GetUser(context)

	// get all notifications for user
	notifications, err := models.GetReceivedNotifications(context.DB, user.ID)
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
	context.Data = struct { Notifications []*models.Notification } { Notifications: notifications, }
}

func RespondNotification(context *system.Context) {
	// parse parameters
	// notificationID := context.Params.GetRequiredId("id")
	// action := context.Params.GetRequiredString("action")


}

func TestNotifications(context *system.Context) {
	// current user
	user := system.GetUser(context)

	// create test notifications
	for i := 0; i < 10; i++ {
		actions := make([]models.NotificationAction, i % 3)
		for j := 0; j < len(actions); j++ {
			actions[j].Name = "Action"
			actions[j].URL = fmt.Sprintf("/notification/respond?action=%d", j)
		}

		notification := &models.Notification {
			SenderID: user.ID, // TODO
			ReceiverID: user.ID,
			ExpiresAt: time.Now().Add(time.Minute),
			Type: models.DefaultNotification,
			//Image: "",
			Message: fmt.Sprintf("Random message %d", i),
			Actions: actions,
			//Data: bson.M {},
		}

		util.Must(notification.Save(context.DB))
	}
}
