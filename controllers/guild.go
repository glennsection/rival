package controllers

import (
	"bloodtales/models"
	"bloodtales/system"
	"bloodtales/util"
	"time"

	"fmt"
	"strings"
	"bytes"

	"gopkg.in/mgo.v2/bson"
)

func handleGuild() {
	handleGameAPI("/guild/create", system.TokenAuthentication, CreateGuild)
	handleGameAPI("/guild/delete", system.TokenAuthentication, DeleteGuild)
	handleGameAPI("/guild/getGuilds", system.TokenAuthentication, GetGuilds)
	handleGameAPI("/guild/getGuildById", system.TokenAuthentication, GetGuildById)
	handleGameAPI("/guild/requestJoin", system.TokenAuthentication, RequestToJoin)
	handleGameAPI("/guild/inviteGuild", system.TokenAuthentication, InviteToGuild)
	handleGameAPI("/guild/addMember", system.TokenAuthentication, AddMember)
	handleGameAPI("/guild/removeMember", system.TokenAuthentication, RemoveMember)
	handleGameAPI("/guild/chat", system.TokenAuthentication, GuildChat)
	handleGameAPI("/guild/shareReplay", system.TokenAuthentication, ShareReplayToGuild)
	handleGameAPI("/guild/guildBattle", system.TokenAuthentication, GuildBattle)
	handleGameAPI("/guild/updateGuildIcon", system.TokenAuthentication, UpdateGuildIcon)
	handleGameAPI("/guild/promote", system.TokenAuthentication, PromoteGuildMember)
	handleGameAPI("/guild/demote", system.TokenAuthentication, DemoteGuildMember)
}

func CreateGuild(context *util.Context) {
	// parse parameters
	name := context.Params.GetRequiredString("name")
	iconId := context.Params.GetRequiredString("iconId")
	description := context.Params.GetString("description", "")
	private := context.Params.GetBool("private", false)

	// get player
	player := GetPlayer(context)

	//Make sure player has enough currency to purchase
	if player.StandardCurrency < models.GuildCreationCost {
		context.Fail("Insufficient funds")
		return
	}
	player.StandardCurrency -= models.GuildCreationCost

	// TODO - make sure player doesn't already own a guild...

	//Make sure that the guild name is unique
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
	if len(guilds) > 0 {
		//Return invalid name
		err := util.NewError("Guild name is already taken. Please choose another")
		util.Must(err)
		return
	}

	// create guild
	guild, err := models.CreateGuild(context, player, name, iconId, description, private)
	util.Must(err)

	SendNotification(context, player, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
}

func DeleteGuild(context *util.Context) {
	player := GetPlayer(context)

	guild, err := models.GetGuildById(context, player.GuildID)
	util.Must(err)

	var memberPlayers []*models.Player
	err = context.DB.C(models.PlayerCollectionName).Find(bson.M{"gd": guild.ID}).All(&memberPlayers)
	util.Must(err)

	// Remove all members except for owner first
	for _, memberPlayer := range memberPlayers {
		if memberPlayer.ID != player.ID {
			RemoveMemberDeleteGuild(context, guild, memberPlayer)
		}
	}

	//Remove owner and delete the guild
	models.RemoveMember(context, player, guild)
}

func GetGuildById(context *util.Context) {
	id := context.Params.GetString("id", "")

	guild, err := models.GetGuildById(context, bson.ObjectIdHex(id))
	util.Must(err)

	context.SetData("guild", guild)
}

func GetGuildByTag(context *util.Context) {
	tag := context.Params.GetString("tag", "")

	guild, err := models.GetGuildByTag(context, tag)
	util.Must(err)

	context.SetData("guild", guild)
}

/// Returns Guild Clients for right now
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

	var guildClients []*models.GuildClient
	for _, guild := range guilds {
		guildClient, err2 := guild.CreateGuildClient(context)
		util.Must(err2)

		guildClients = append(guildClients, guildClient)
	}

	// result
	context.SetData("guilds", guildClients)
}

func RequestToJoin(context *util.Context) {
	guildTag := context.Params.GetRequiredString("guildTag")
	playerTag := context.Params.GetRequiredString("playerTag")

	fmt.Printf("Inside of RequestToJoin")
	player, err1 := models.GetPlayerByTag(context, playerTag)
	util.Must(err1)

	// guild
	guild, err := models.GetGuildByTag(context, guildTag)
	util.Must(err)

	var buffer bytes.Buffer
	buffer.WriteString(player.Name)
	buffer.WriteString(" Request To Join the Guild")

	requestData := map[string]interface{}{"requestType": "RequestToJoin", "playerTag": playerTag}

	SendGuildChatNotification(context, "GuildChat", buffer.String(), models.PlayerDataMask_Guild, "Accept", "accept", "Decline", "decline", requestData, time.Now().Add(time.Hour*time.Duration(168)), guild, false)
}

func InviteToGuild(context *util.Context) {
	guildTag := context.Params.GetRequiredString("guildTag")
	playerTag := context.Params.GetRequiredString("playerTag")

	fmt.Printf("Inside of InviteToGuild")
	receiverPlayer, err1 := models.GetPlayerByTag(context, playerTag)
	util.Must(err1)

	// guild
	guild, err := models.GetGuildByTag(context, guildTag)
	util.Must(err)

	//TODO Build message with username and guild name
	inviteMessage := []string{GetPlayer(context).Name, "Invited you to Guild:", guild.Name}
	inviteData := map[string]interface{}{"guildTag": guildTag}

	SendNotification(context, receiverPlayer, "GuildInvite", strings.Join(inviteMessage, " "), models.PlayerDataMask_Guild, "Accept", "accept", "Decline", "decline", inviteData, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
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

	//Chat message
	var buffer bytes.Buffer
	buffer.WriteString(player.Name)
	buffer.WriteString(" Has Joined the Guild")

	SendGuildChatNotification(context, "GuildChat", buffer.String(), models.PlayerDataMask_Guild, "Accept", "accept", "Decline", "decline", nil, time.Now().Add(time.Hour*time.Duration(168)), nil, false)
	SendGuildChatNotification(context, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
}

func RemoveMember(context *util.Context) {
	guildTag := context.Params.GetRequiredString("guildTag")
	playerTag := context.Params.GetRequiredString("playerTag")

	var player *models.Player
	if playerTag == "" {
		player = GetPlayer(context)
	} else {
		var err1 error
		player, err1 = models.GetPlayerByTag(context, playerTag)
		util.Must(err1)
	}

	// guild
	guild, err := models.GetGuildByTag(context, guildTag)
	//guild, err := models.GetGuildById(context, bson.ObjectIdHex(tag))
	util.Must(err)

	//Notify everyone they have left
	var buffer bytes.Buffer
	buffer.WriteString(player.Name)
	buffer.WriteString(" Has Left the Guild")
	
	SendGuildChatNotification(context, "GuildChat", buffer.String(), models.PlayerDataMask_Guild, "Accept", "accept", "Decline", "decline", nil, time.Now().Add(time.Hour*time.Duration(168)), nil, false)

	err2 := models.RemoveMember(context, player, guild)
	util.Must(err2)

	SendNotification(context, player, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
	SendGuildChatNotification(context, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
}

func RemoveMemberDeleteGuild(context *util.Context, guild *models.Guild, player *models.Player) {

	err2 := models.RemoveMember(context, player, guild)
	util.Must(err2)

	SendNotification(context, player, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
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

	//Notify everyone they have been promoted
	var buffer bytes.Buffer
	buffer.WriteString(player.Name)
	buffer.WriteString(" Has been Promoted to ")
	buffer.WriteString(models.GetGuildRoleName(player.GuildRole))
	
	SendGuildChatNotification(context, "GuildChat", buffer.String(), models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(168)), nil, false)

	SendGuildChatNotification(context, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
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

	//Notify everyone they have been promoted
	var buffer bytes.Buffer
	buffer.WriteString(player.Name)
	buffer.WriteString(" Has been Demoted to ")
	buffer.WriteString(models.GetGuildRoleName(player.GuildRole))
	
	SendGuildChatNotification(context, "GuildChat", buffer.String(), models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(168)), nil, false)

	SendGuildChatNotification(context, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
}

func SendNotification(context *util.Context, receiverPlayer *models.Player, notificationType string, message string, dirtyMask util.Bits, acceptName string, acceptAction string, declineName string,
	declineAction string, data map[string]interface{}, expiresAt time.Time, guild *models.Guild, setDirty bool) {
	//notificationType := "GuildChat"
	// sending player
	senderPlayer := GetPlayer(context)
	senderPlayerClient, err := senderPlayer.GetPlayerClient(context)
	util.Must(err)

	// create notification
	notification := &models.Notification{
		SenderID:   senderPlayer.ID,
		ReceiverID: receiverPlayer.ID,
		Guild:      true, // TODO - guild chat based on "channel"
		ExpiresAt:  time.Now().Add(time.Hour * time.Duration(168)),
		Type:       notificationType,
		Message:    message,
		SenderName: senderPlayerClient.Name,
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

	//Marshall new player data
	receiverPlayer.SetDirty(dirtyMask)
	playerData := receiverPlayer.MarshalDirty(context)

	// notify receiver
	socketData := map[string]interface{}{"notification": notification, "senderPlayer": senderPlayerClient, "playerData": playerData, "playerDataMask": dirtyMask}
	system.SocketSend(context, receiverPlayer.UserID, notificationType, socketData)
}

func SendGuildChatNotification(context *util.Context, notificationType string, message string, dirtyMask util.Bits, acceptName string, acceptAction string, declineName string,
	declineAction string, data map[string]interface{}, expiresAt time.Time, guild *models.Guild, setDirty bool) {
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

	if guild == nil {
		guild2, err := models.GetGuildById(context, player.GuildID)
		guild = guild2
		util.Must(err)
	}

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

		if setDirty {
			memberPlayer.SetDirty(dirtyMask)

			playerData := memberPlayer.MarshalDirty(context)

			// notify receiver
			socketDataDirty := map[string]interface{}{"notification": notification, "player": playerClient, "playerData": playerData, "playerDataMask": dirtyMask}
			system.SocketSend(context, memberPlayer.UserID, notificationType, socketDataDirty)
		} else {
			// notify receiver
			socketData := map[string]interface{}{"notification": notification, "player": playerClient}
			system.SocketSend(context, memberPlayer.UserID, notificationType, socketData)
		}
	}
}

func GuildChat(context *util.Context) {
	// parse parameters
	//channel := context.Params.GetString("channel", "")
	message := context.Params.GetRequiredString("message")

	SendGuildChatNotification(context, "GuildChat", message, models.PlayerDataMask_Guild, "Accept", "accept", "Decline", "decline", nil, time.Now().Add(time.Hour*time.Duration(168)), nil, false)
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

	data := bson.M{"requestType": "ShareReplay", "replayInfo": replayInfo}

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
		system.SocketSend(context, memberPlayer.UserID, notificationType, socketData)
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
	arenaName := context.Params.GetRequiredString("arenaName")
	//tag := context.Params.GetRequiredString("tag")

	// generate Room ID
	roomID := util.GenerateUUID()
	//message := fmt.Sprintf("Battle Request from: %s", models.GetUserName(context, context.UserID))
	data := map[string]interface{}{
		"roomId": roomID,
		"arenaName": arenaName,
	}
	expiresAt := time.Now().Add(time.Hour)

	fmt.Printf("Creating Guild Battle Notification")

	SendGuildChatNotification(context, "GuildBattle", message, models.PlayerDataMask_Guild, "Accept", "accept", "Decline", "decline", data, expiresAt, nil, false)
	//sendFriendNotification(context, tag, "FriendBattle", image, message, "Battle", "accept", "Decline", "decline", data, expiresAt)

	context.SetData("roomId", roomID)
	context.SetData("arenaName", arenaName)
}

func respondGuildBattle(context *util.Context, notification *models.Notification, action string) {
	if action == "accept" {
		// create private match
		roomID := notification.Data["roomId"].(string)
		arenaName := notification.Data["arenaName"].(string)
		player := GetPlayer(context)
		_, err := models.StartPrivateMatch(context, notification.SenderID, player.ID, models.MatchRanked, roomID, arenaName)
		util.Must(err)
	}
}

func respondGuildJoinRequest(context *util.Context, notification *models.Notification, action string) {

	if action == "accept" {
		//TODO Do we need to send a popup or something here?
		//TODO AddMember done in here instead of call from client
		//SendGuildChatNotification(context, "UpdateGuildInfo", "", models.PlayerDataMask_Guild, "", "", "", "", nil, time.Now().Add(time.Hour*time.Duration(1)), guild, true)
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
