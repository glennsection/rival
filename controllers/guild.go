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
	handleGameAPI("/guild/delete", system.TokenAuthentication, DeleteGuild)
	handleGameAPI("/guild/getGuilds", system.TokenAuthentication, GetGuilds)
	handleGameAPI("/guild/getGuildById", system.TokenAuthentication, GetGuildById)
	handleGameAPI("/guild/addMember", system.TokenAuthentication, AddMember)
	handleGameAPI("/guild/removeMember", system.TokenAuthentication, RemoveMember)
	handleGameAPI("/guild/chat", system.TokenAuthentication, GuildChat)
	handleGameAPI("/guild/shareReplay", system.TokenAuthentication, ShareReplayToGuild)
	handleGameAPI("/guild/guildBattle", system.TokenAuthentication, GuildBattle)
	handleGameAPI("/guild/updateGuildIcon",  system.TokenAuthentication, UpdateGuildIcon)
	handleGameAPI("/guild/promote",  system.TokenAuthentication, PromoteGuildMember)
	handleGameAPI("/guild/demote",  system.TokenAuthentication, DemoteGuildMember)
}

func CreateGuild(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")
	iconId := context.Params.GetRequiredString("iconId")

	// get player
	player := GetPlayer(context)

	// TODO - make sure player doesn't already own a guild...

	// TODO - make sure that the guild name is unique
	var guilds []*models.Guild
	if name != "" {
		util.Must(context.DB.C(models.GuildCollectionName).Find(bson.M{
			"nm": bson.M{
				"$regex": bson.RegEx{
					Pattern: fmt.Sprintf("%s", name),
					Options: "i",
				},
			},
		}).All(&guilds))
	}
	fmt.Printf("Length of Guilds with same name: %d", len(guilds))
	if (len(guilds) > 0) {
		//Return invalid name
		err := util.NewError("Guild name is already taken. Please choose another")
		util.Must(err)
		return
	}


	// create guild
	_, err := models.CreateGuild(context, player, name, iconId)
	util.Must(err)
}

func DeleteGuild(context *util.Context) {
	player := GetPlayer(context)

	guild,err := models.GetGuildById(context, player.GuildID)
	util.Must(err)

	err2 := guild.Delete(context)
	util.Must(err2)
}

func GetGuildById(context *util.Context) {
	id := context.Params.GetString("id", "")

	guild,err := models.GetGuildById(context, bson.ObjectIdHex(id))
	util.Must(err)

	context.SetData("guild", guild)
}

func GetGuildByTag(context *util.Context) {
	tag := context.Params.GetString("tag", "")

	guild,err := models.GetGuildByTag(context, tag)
	util.Must(err)

	context.SetData("guild", guild)
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
	guildTag := context.Params.GetRequiredString("guildTag")
	tag := context.Params.GetRequiredString("tag")

	//player := GetPlayer(context)
	player, err1 := models.GetPlayerByTag(context, tag)
	util.Must(err1)

	// guild
	guild, err := models.GetGuildByTag(context, guildTag)
	util.Must(err)

	err2 := models.AddMember(context, player, guild)
	util.Must(err2)
}

func RemoveMember(context *util.Context) {
	guildTag := context.Params.GetRequiredString("guildTag")
	playerTag := context.Params.GetRequiredString("playerTag")

	//player := GetPlayer(context)

	// guild
	guild, err := models.GetGuildByTag(context, guildTag)
	//guild, err := models.GetGuildById(context, bson.ObjectIdHex(tag))
	util.Must(err)

	player, err1 := models.GetPlayerByTag(context, playerTag)
	util.Must(err1)

	err2 := models.RemoveMember(context, player, guild)
	util.Must(err2)
}

func PromoteGuildMember(context *util.Context) {
	guildTag := context.Params.GetRequiredString("guildTag")
	playerTag := context.Params.GetRequiredString("playerTag")

	guild, err := models.GetGuildByTag(context, guildTag)
	//guild, err := models.GetGuildById(context, bson.ObjectIdHex(tag))
	util.Must(err)

	player, err1 := models.GetPlayerByTag(context, playerTag)
	util.Must(err1)

	err2 := models.PromoteGuildUser(context, player, guild)
	util.Must(err2)

	context.SetData("guildRole", models.GetGuildRoleName(player.GuildRole))
}

func DemoteGuildMember(context *util.Context) {
	guildTag := context.Params.GetRequiredString("guildTag")
	playerTag := context.Params.GetRequiredString("playerTag")

	guild, err := models.GetGuildByTag(context, guildTag)
	//guild, err := models.GetGuildById(context, bson.ObjectIdHex(tag))
	util.Must(err)

	player, err1 := models.GetPlayerByTag(context, playerTag)
	util.Must(err1)

	err2 := models.DemoteGuildUser(context, player, guild)
	util.Must(err2)

	context.SetData("guildRole", models.GetGuildRoleName(player.GuildRole))
}

func SendGuildChatNotification(context *util.Context, notificationType string, message string, acceptName string, acceptAction string, declineName string, declineAction string, data map[string]interface{}, expiresAt time.Time) {
	//notificationType := "GuildChat"

	fmt.Printf("Inside send guild chat notification")
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

	fmt.Printf("Creating Guild Battle Notification")

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

func UpdateGuildIcon(context *util.Context) {
	iconId := context.Params.GetRequiredString("iconId")

	player := GetPlayer(context)

	guild, err := models.GetGuildById(context, player.GuildID)
	util.Must(err)

	err2 := models.UpdateGuildIcon(context, player, guild, iconId)
	util.Must(err2)
}
