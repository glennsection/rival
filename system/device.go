package system

import (
	"errors"

	"bloodtales/models"
)

func (context *Context) authenticateDevice(required bool) (err error) {
	// parse parameters
	uuid := context.Params.GetString("uuid", "")
	tag := context.Params.GetString("tag", "")
	token := context.Params.GetString("debug", "")

	if tag != "" {
		// login using player tag
		if token == context.Config.Authentication.DebugToken {
			context.User, err = models.GetUserByTag(context.DB, tag)
		}
	} else {
		// handle identifier as UUID
		if uuid != "" {
			// find user with UUID
			context.User, err = models.GetUserByUUID(context.DB, uuid)
			if required && context.User == nil {
				// generate unique player tag
				tag := GenerateTag()

				// insert new user
				context.User, err = models.InsertUserWithUUID(context.DB, uuid, tag)
				if err != nil {
					panic(err)
				}

				// insert new player
				player := models.CreatePlayer(context.User.ID, "")
				err = player.Update(context.DB)
				if err != nil {
					panic(err)
				}
			}

			err = context.AppendAuthToken()
		}
	}

	if required && context.User == nil {
		err = errors.New("Unauthorized user")
	}
	return
}
