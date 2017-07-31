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
	handleGameAPI("/guild/chat", system.TokenAuthentication, GuildChat)
	handleGameAPI("/guild/shareReplay", system.TokenAuthentication, ShareReplayToGuild)
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

func SendGuildChatNotification(context *util.Context, channel string, message string) {
	notificationType := "GuildChat"

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
	channel := context.Params.GetString("channel", "")
	message := context.Params.GetRequiredString("message")

	SendGuildChatNotification(context, channel, message)
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
	message  := context.Params.GetRequiredString("message")

	SendReplayGuildNotification(context, replayInfoId, message)
}
