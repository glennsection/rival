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
	handleGameAPI("/friends/add", system.TokenAuthentication, AddFriend)
}

func GetFriends(context *util.Context) {
	// current player
	player := GetPlayer(context)

	var client []*models.PlayerClient

	// get all friends for player
	friends, err := models.GetFriendsByPlayerId(context.DB, player.ID, false)
	util.Must(err)

	if friends != nil {
		// get friend players
		var friendPlayers []*models.Player
		util.Must(context.DB.C(models.PlayerCollectionName).Find(bson.M {
			"_id": bson.M { "$in": friends.FriendIDs, },
		}).All(&friendPlayers))

		// create client array
		for _, friendPlayer := range friendPlayers {
			playerClient, err := friendPlayer.CreatePlayerClient(context.DB)
			util.Must(err)

			client = append(client, playerClient)
		}
	}

	// result
	context.SetData("friends", client)
}

func AddFriend(context *util.Context) {
	// parse parameters
	tag := context.Params.GetRequiredString("tag")

	// current player
	player := GetPlayer(context)

	// friend player
	friendUser, err := models.GetUserByTag(context.DB, tag)
	util.Must(err)
	friendPlayer, err := models.GetPlayerByUser(context.DB, friendUser.ID)
	util.Must(err)

	// create notification
	notification := &models.Notification {
		SenderID: player.ID,
		ReceiverID: friendPlayer.ID,
		Guild: false,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(168)),
		Type: "FriendRequest",
		//Image: "",
		Message: fmt.Sprintf("Friend Request from: %s", GetPlayerName(context, player.ID)),
		Actions: []models.NotificationAction {
			models.NotificationAction {
				Name: "Accept",
				Value: "accept",
			},
			models.NotificationAction {
				Name: "Decline",
				Value: "decline",
			},
		},
		//Data: bson.M {},
	}
	util.Must(notification.Save(context.DB))

	// notify receiver
	system.SocketSend(friendUser.ID, "FriendRequested", map[string]interface{} { "notification": notification })
}

func AcceptFriend(context *util.Context, senderID bson.ObjectId, receiverID bson.ObjectId) {
	// get sender friends
	var senderFriends *models.Friends
	senderFriends, err := models.GetFriendsByPlayerId(context.DB, senderID, true)
	util.Must(err)

	// append sender friends
	senderFriends.FriendIDs = append(senderFriends.FriendIDs, receiverID)
	err = senderFriends.Save(context.DB)
	util.Must(err)

	// get receiver friends
	var receiverFriends *models.Friends
	receiverFriends, err = models.GetFriendsByPlayerId(context.DB, receiverID, true)
	util.Must(err)

	// append receiver friends
	receiverFriends.FriendIDs = append(receiverFriends.FriendIDs, senderID)
	err = receiverFriends.Save(context.DB)
	util.Must(err)

	// notify sender
	senderUserID := GetUserIdByPlayerId(context, senderID)
	system.SocketSend(senderUserID, "FriendAccepted", nil)
}