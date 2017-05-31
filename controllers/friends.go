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

	util.Must((&models.Notification {
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
	}).Save(context.DB))
}
