package system

import (
	"errors"
	"strings"

	"bloodtales/config"
	"bloodtales/util"
	"bloodtales/models"
)

func authenticateCredentials(context *util.Context, required bool) (err error) {
	// parse tag parameters
	tag := context.Params.GetString("tag", "")
	token := context.Params.GetString("debug", "")

	var user *models.User = nil

	if tag != "" {
		// login using player tag
		if token == config.Config.Authentication.DebugToken {
			user, err = models.GetUserByTag(context, tag)
			util.MustIgnoreNotFound(err)

			if user != nil {
				SetUser(context, user)
			}
		}
	} else {
		// parse credentials parameters
		credentialsParameter := context.Params.GetString("credentials", "")
		var credentials []models.Credential = nil

		// login using credentials
		if credentialsParameter != "" {
			// parse credentials
			credentialsPairs := strings.Split(credentialsParameter, ",")

			// build credentials
			credentials := make([]models.Credential, len(credentialsPairs))
			for i, credentialsPair := range credentialsPairs {
				credentialsParts := strings.Split(credentialsPair, ":")
				credentials[i] = models.Credential { Provider: credentialsParts[0], ID: credentialsParts[1] }
			}

			// find user with credentials
			user, err = models.GetUserByCredentials(context, credentials)
		}

		// check if no user was found with credentials
		if user == nil {
			// parse device parameters
			uuid := context.Params.GetString("uuid", "")

			// find user with device UUID
			user, err = models.GetUserByDevice(context, uuid)
			util.MustIgnoreNotFound(err)

			// check if we need to create or update a new user
			if user != nil {
				if credentials != nil {
					// add credentials to user
					util.Must(user.AppendCredentials(context, credentials))
				}
			} else if required {
				// insert new user
				user, err = models.InsertUserWithDevice(context, uuid, credentials)
				util.Must(err)

				// inform controller that we need to reset this player
				context.Params.Set("reset", true)
			}
		}
	}

	if user != nil {
		SetUser(context, user)

		err = IssueAuthToken(context)
	} else if required {
		err = errors.New("Unauthorized user")
	}
	return
}
