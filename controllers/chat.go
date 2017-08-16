package controllers

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handleChat() {
	handleGameAPI("/chat", system.TokenAuthentication, Chat)
}

func sendChatNotification(context *util.Context, channel string, message string) {
	notificationType := "Chat"

	// sending player
	player := GetPlayer(context)
	playerClient, err := player.GetPlayerClient(context)
	util.Must(err)

	// friend player (TODO - player to player chat?)
	// friendUser, err := models.GetUserByTag(context, tag)
	// util.Must(err)
	// friendPlayer, err := models.GetPlayerByUser(context, friendUser.ID)
	// util.Must(err)
	// receiverID := friendPlayer.ID
	var receiverID bson.ObjectId = bson.ObjectId("")
	var receiverUserID bson.ObjectId = bson.ObjectId("")

	// create notification
	notification := &models.Notification {
		SenderID: player.ID,
		ReceiverID: receiverID,
		Guild: false, // TODO - guild chat based on "channel"
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(168)),
		Type: notificationType,
		//Image: image,
		Message: message,
	}
	util.Must(notification.Save(context))

	// notify receiver
	socketData := map[string]interface{} { "notification": notification, "player": playerClient }
	system.SocketSend(context, receiverUserID, notificationType, socketData)
}

func prepareChatNotification(context *util.Context, notification *models.Notification) {
	// get sender player
	player, err := models.GetPlayerById(context, notification.SenderID)
	util.Must(err)

	// create sender player info
	var playerClient *models.PlayerClient
	playerClient, err = player.GetPlayerClient(context)
	util.Must(err)

	// add sender player info
	notification.Data = bson.M { "player": playerClient }
}

func Chat(context *util.Context) {
	// parse parameters
	channel := context.Params.GetString("channel", "")
	message := context.Params.GetRequiredString("message")

	sendChatNotification(context, channel, message)
}
