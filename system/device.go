package system

import (
	"errors"

	"bloodtales/config"
	"bloodtales/util"
	"bloodtales/models"
)

func authenticateDevice(context *util.Context, required bool) (err error) {
	// parse parameters
	uuid := context.Params.GetString("uuid", "")
	tag := context.Params.GetString("tag", "")
	token := context.Params.GetString("debug", "")

	var user *models.User = nil

	if tag != "" {
		// login using player tag
		if token == config.Config.Authentication.DebugToken {
			user, err = models.GetUserByTag(context.DB, tag)

			if user != nil {
				SetUser(context, user)
			}
		}
	} else {
		// handle identifier as UUID
		if uuid != "" {
			// find user with UUID
			user, err = models.GetUserByUUID(context.DB, uuid)

			if required && user == nil {
				// generate unique player tag
				tag := util.GenerateTag()

				// insert new user
				user, err = models.InsertUserWithUUID(context.DB, uuid, tag)
				util.Must(err)

				// insert new player
				player := models.CreatePlayer(user.ID)
				util.Must(player.Save(context.DB))
			}
		}
	}

	if user != nil {
		SetUser(context, user)

		err = AppendAuthToken(context)
	} else if required {
		err = errors.New("Unauthorized user")
	}
	return
}
