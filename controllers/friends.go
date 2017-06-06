package controllers

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/models"
)

func handleFriends() {
	handleGameAPI("/friends/get", system.TokenAuthentication, GetFriends)
	handleGameAPI("/friends/add", system.TokenAuthentication, FriendRequest)
	handleGameAPI("/friends/battle", system.TokenAuthentication, FriendBattle)
}

func sendFriendNotification(context *util.Context, tag string, notificationType string, image string, message string, acceptName string, acceptAction string, declineName string, declineAction string, data map[string]interface{}, expiresAt time.Time) {
	// sending player
	player := GetPlayer(context)
	playerClient, err := player.GetPlayerClient(context)
	util.Must(err)

	// friend player
	friendUser, err := models.GetUserByTag(context, tag)
	util.Must(err)
	friendPlayer, err := models.GetPlayerByUser(context, friendUser.ID)
	util.Must(err)

	// create notification
	notification := &models.Notification {
		SenderID: player.ID,
		ReceiverID: friendPlayer.ID,
		Guild: false,
		ExpiresAt: expiresAt,
		Type: notificationType,
		Image: image,
		Message: message,
		Actions: []models.NotificationAction {
			models.NotificationAction {
				Name: acceptName,
				Value: acceptAction,
			},
			models.NotificationAction {
				Name: declineName,
				Value: declineAction,
			},
		},
		Data: data,
	}
	util.Must(notification.Save(context))

	// notify receiver
	socketData := map[string]interface{} { "notification": notification, "player": playerClient }
	for key, value := range data {
		socketData[key] = value
	}
	system.SocketSend(friendUser.ID, notificationType, socketData)
}

func prepareFriendNotification(context *util.Context, notification *models.Notification) {
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

func GetFriends(context *util.Context) {
	// current player
	player := GetPlayer(context)

	var client []*models.PlayerClient

	// get all friends for player
	friends, err := models.GetFriendsByPlayerId(context, player.ID, false)
	util.Must(err)

	if friends != nil {
		// get friend players
		var friendPlayers []*models.Player
		util.Must(context.DB.C(models.PlayerCollectionName).Find(bson.M {
			"_id": bson.M { "$in": friends.FriendIDs, },
		}).All(&friendPlayers))

		// create client array
		for _, friendPlayer := range friendPlayers {
			playerClient, err := friendPlayer.GetPlayerClient(context)
			util.Must(err)

			client = append(client, playerClient)
		}
	}

	// result
	context.SetData("friends", client)
}

func FriendRequest(context *util.Context) {
	// parse parameters
	tag := context.Params.GetRequiredString("tag")

	image := ""
	message := fmt.Sprintf("Friend Request from: %s", GetUserName(context, context.UserID))
	expiresAt := time.Now().Add(time.Hour * time.Duration(168))

	sendFriendNotification(context, tag, "FriendRequest", image, message, "Accept", "accept", "Decline", "decline", nil, expiresAt)
}

func acceptFriendRequest(context *util.Context, senderID bson.ObjectId, receiverID bson.ObjectId) {
	// get sender friends
	var senderFriends *models.Friends
	senderFriends, err := models.GetFriendsByPlayerId(context, senderID, true)
	util.Must(err)

	// append sender friends
	senderFriends.FriendIDs = append(senderFriends.FriendIDs, receiverID)
	err = senderFriends.Save(context)
	util.Must(err)

	// get receiver friends
	var receiverFriends *models.Friends
	receiverFriends, err = models.GetFriendsByPlayerId(context, receiverID, true)
	util.Must(err)

	// append receiver friends
	receiverFriends.FriendIDs = append(receiverFriends.FriendIDs, senderID)
	err = receiverFriends.Save(context)
	util.Must(err)
}

func FriendBattle(context *util.Context) {
	// parse parameters
	tag := context.Params.GetRequiredString("tag")

	// generate Room ID
	roomID := util.GenerateUUID()

	image := ""
	message := fmt.Sprintf("Battle Request from: %s", GetUserName(context, context.UserID))
	data := map[string]interface{} {
		"roomId": roomID,
	}
	expiresAt := time.Now().Add(time.Hour)

	sendFriendNotification(context, tag, "FriendBattle", image, message, "Battle", "accept", "Decline", "decline", data, expiresAt)

	context.SetData("roomId", roomID)
}
