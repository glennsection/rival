package controllers

import (
	"bloodtales/config"
	"bloodtales/system"
	"bloodtales/util"
)

func handleUser() {
	handleGameAPI("/connect", system.NoAuthentication, UserConnect)
	handleGameAPI("/login", system.DeviceAuthentication, UserLogin)
	handleGameAPI("/reauth", system.TokenAuthentication, UserReauth)
	handleGameAPI("/logout", system.TokenAuthentication, UserLogout)
}

func UserConnect(context *util.Context) {
	// parse parameters
	version := context.Params.GetRequiredString("version")

	// check version (major and minor)
	compatibility := util.CompareVersion(config.Config.Platform.Version, version, 2)
	switch compatibility {

	case -1:
		context.Fail("Client version is behind server.  Please update client!")

	case 1:
		context.Fail("Client version is ahead of server.  Please update server!")

	}

	// update client values
	context.Client.Version = version
	context.Client.Save()
}

func UserLogin(context *util.Context) {
	// analytics tracking
	InsertTracking(context, "login", nil, 0)

	// respond with player data
	FetchPlayer(context)
}

func UserReauth(context *util.Context) {
	// issue new auth token
	system.IssueAuthToken(context)
}

func UserLogout(context *util.Context) {
	// clear auth token
	system.ClearAuthToken(context)
}
