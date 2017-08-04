package controllers

import (
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
	"time"

	"fmt"

	"gopkg.in/mgo.v2/bson"
)

func handleGuild() {
	handleGameAPI("/guild/create", system.TokenAuthentication, CreateGuild)
	handleGameAPI("/guild/getGuilds", system.TokenAuthentication, GetGuilds)
	handleGameAPI("/guild/addMember", system.TokenAuthentication, AddMember)
	handleGameAPI("/guild/removeMember", system.TokenAuthentication, RemoveMember)
	handleGameAPI("/guild/chat", system.TokenAuthentication, GuildChat)
	handleGameAPI("/guild/shareReplay", system.TokenAuthentication, ShareReplayToGuild)
	handleGameAPI("/guild/guildBattle", system.TokenAuthentication, GuildBattle)
}

func CreateGuild(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")

	// get player
	player := GetPlayer(context)

	// TODO - make sure player doesn't already own a guild...

	// create guild
	_, err := models.CreateGuild(context, player, name)
	util.Must(err)
}

func GetGuilds(context *util.Context) {
	// parse parameters
	name := context.Params.GetString("name", "")

	var guilds []*models.Guild
	if name != "" {
		util.Must(context.DB.C(models.GuildCollectionName).Find(bson.M{
			"nm": bson.M{
				"$regex": bson.RegEx{
					Pattern: fmt.Sprintf(".*%s.*", name),
					Options: "i",
				},
			},
		}).All(&guilds))
	} else {
		util.Must(context.DB.C(models.GuildCollectionName).Find(nil).All(&guilds))
	}
	//util.Must(context.DB.C(models.GuildCollectionName).Find(bson.M{"nm": bson.M{"$eq": "MikeMasterGuild"}}).All(&guilds))

	// result
	context.SetData("guilds", guilds)
}

func AddMember(context *util.Context) {
	tag := context.Params.GetRequiredString("tag")

	player := GetPlayer(context)

	// guild
	guild, err := models.GetGuildByTag(context, tag)
	util.Must(err)

	err2 := models.AddMember(context, player, guild)
	util.Must(err2)
}

func RemoveMember(context *util.Context) {
	tag := context.Params.GetRequiredString("tag")

	player := GetPlayer(context)

	// guild
	guild, err := models.GetGuildByTag(context, tag)
	//guild, err := models.GetGuildById(context, bson.ObjectIdHex(tag))
	util.Must(err)

	err2 := models.RemoveMember(context, player, guild)
	util.Must(err2)
}

func SendGuildChatNotification(context *util.Context, notificationType string, message string, acceptName string, acceptAction string, declineName string, declineAction string, data map[string]interface{}, expiresAt time.Time) {
	//notificationType := "GuildChat"

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
	//var receiverID bson.ObjectId = bson.ObjectId("")
	//var receiverUserID bson.ObjectId = bson.ObjectId("")

	guild, err := models.GetGuildById(context, player.GuildID)
	util.Must(err)

	var memberPlayers []*models.Player
	err = context.DB.C(models.PlayerCollectionName).Find(bson.M{"gd": guild.ID}).All(&memberPlayers)
	util.Must(err)

	// create notification
	notification := &models.Notification{
		SenderID:   player.ID,
		ReceiverID: player.GuildID,
		Guild:      true, // TODO - guild chat based on "channel"
		ExpiresAt:  time.Now().Add(time.Hour * time.Duration(168)),
		Type:       notificationType,
		Message:    message,
		SenderName: playerClient.Name,
		Actions: []models.NotificationAction{
			models.NotificationAction{
				Name:  acceptName,
				Value: acceptAction,
			},
			models.NotificationAction{
				Name:  declineName,
				Value: declineAction,
			},
		},
		Data: data,
	}
	util.Must(notification.Save(context))

	for _, memberPlayer := range memberPlayers {
		// notify receiver
		socketData := map[string]interface{}{"notification": notification, "player": playerClient}
		system.SocketSend(memberPlayer.UserID, notificationType, socketData)
	}
}

func GuildChat(context *util.Context) {
	// parse parameters
	//channel := context.Params.GetString("channel", "")
	message := context.Params.GetRequiredString("message")

	SendGuildChatNotification(context, "GuildChat", message, "Accept", "accept", "Decline", "decline", nil, time.Now().Add(time.Hour*time.Duration(168)))
}

func SendReplayGuildNotification(context *util.Context, replayInfoId string, message string) {
	notificationType := "GuildChat"
	// sending player
	player := GetPlayer(context)
	playerClient, err := player.GetPlayerClient(context)
	util.Must(err)

	guild, err := models.GetGuildById(context, player.GuildID)
	util.Must(err)

	var memberPlayers []*models.Player
	err = context.DB.C(models.PlayerCollectionName).Find(bson.M{"gd": guild.ID}).All(&memberPlayers)
	util.Must(err)

	fmt.Println("infoID:", replayInfoId)
	replayInfo, err := models.GetReplayInfoById(context, bson.ObjectIdHex(replayInfoId))
	util.Must(err)

	data := bson.M{"replayInfo": replayInfo}

	// create notification
	notification := &models.Notification{
		SenderID:   player.ID,
		ReceiverID: player.GuildID,
		Guild:      true, // TODO - guild chat based on "channel"
		ExpiresAt:  time.Now().Add(time.Hour * time.Duration(168)),
		Type:       notificationType,
		Message:    message,
		SenderName: playerClient.Name,
		Data:       data,
	}
	util.Must(notification.Save(context))

	for _, memberPlayer := range memberPlayers {
		// notify receiver
		socketData := map[string]interface{}{"notification": notification, "player": playerClient}
		system.SocketSend(memberPlayer.UserID, notificationType, socketData)
	}
}

func ShareReplayToGuild(context *util.Context) {
	replayInfoId := context.Params.GetRequiredString("infoId")
	message := context.Params.GetRequiredString("message")

	SendReplayGuildNotification(context, replayInfoId, message)
}

func GuildBattle(context *util.Context) {
	// parse parameters
	message := context.Params.GetRequiredString("message")
	//tag := context.Params.GetRequiredString("tag")

	// generate Room ID
	roomID := util.GenerateUUID()
	//message := fmt.Sprintf("Battle Request from: %s", models.GetUserName(context, context.UserID))
	data := map[string]interface{}{
		"roomId": roomID,
	}
	expiresAt := time.Now().Add(time.Hour)

	SendGuildChatNotification(context, "GuildBattle", message, "Accept", "accept", "Decline", "decline", data, expiresAt)
	//sendFriendNotification(context, tag, "FriendBattle", image, message, "Battle", "accept", "Decline", "decline", data, expiresAt)

	context.SetData("roomId", roomID)
}

func respondGuildBattle(context *util.Context, notification *models.Notification, action string) {
	if action == "accept" {
		// create private match
		roomID := notification.Data["roomId"].(string)
		player := GetPlayer(context)
		_, err := models.StartPrivateMatch(context, notification.SenderID, player.ID, models.MatchRanked, roomID)
		util.Must(err)
	}
}
