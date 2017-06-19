package system

import (
	"errors"

	"bloodtales/config"
	"bloodtales/util"
	"bloodtales/models"
)

func authenticateDevice(context *util.Context, required bool) (err error) {
	// parse parameters
	// TODO - should be passing multiple credentials (Provider, ID), and processing the entirety
	credentialProvider := "UUID"
	credentialID := context.Params.GetString("uuid", "")

	tag := context.Params.GetString("tag", "")
	token := context.Params.GetString("debug", "")

	var user *models.User = nil

	if tag != "" {
		// login using player tag
		if token == config.Config.Authentication.DebugToken {
			user, err = models.GetUserByTag(context, tag)

			if user != nil {
				SetUser(context, user)
			}
		}
	} else {
		// login using credentials
		if credentialID != "" {
			// build credentials
			credentials := []models.Credential {
				models.Credential { Provider: credentialProvider, ID: credentialID },
			}

			// find user with credentials
			user, err = models.GetUserByCredentials(context, credentials)

			if required && user == nil {
				// insert new user
				user, err = models.InsertUserWithCredentials(context, credentials)
				util.Must(err)

				// insert new player
				var player *models.Player
				player, err = models.CreatePlayer(user.ID)
				util.Must(err)

				util.Must(player.Save(context))
			}
		}
	}

	if user != nil {
		SetUser(context, user)

		err = issueAuthToken(context)
	} else if required {
		err = errors.New("Unauthorized user")
	}
	return
}
