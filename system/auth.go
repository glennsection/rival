package system

import (
	"fmt"
	"errors"

	"bloodtales/util"
	"bloodtales/models"
)

// auth type enum
type AuthenticationType int
const (
	NoAuthentication AuthenticationType = iota
	AnyAuthentication
	DeviceAuthentication
	PasswordAuthentication
	TokenAuthentication
)

func init() {
	util.AddTemplateFunc("authenticated", Authenticated)
}

func SetUser(context *util.Context, user *models.User) {
	context.UserID = user.ID
	context.Params.Set("_user", user)
}

func GetUser(context *util.Context) *models.User {
	if user, ok := context.Params.Get("_user").(*models.User); ok {
		return user
	}
	return nil
}

// check if context is authenticated
func Authenticated(context *util.Context) bool {
	return GetUser(context) != nil
}

// basic username/password auth
func authenticatePassword(context *util.Context, required bool) (err error) {
	// parse login parameters
	username, password := context.Params.GetString("username", ""), context.Params.GetString("password", "")

	// authenticate user
	if username != "" && password != "" {
		var user *models.User
		user, err = models.LoginUser(context, username, password)
		if user == nil {
			err = errors.New(fmt.Sprintf("Invalid authentication information for username: %v (%v)", username, err))
			return
		}

		// set user in context
		SetUser(context, user)

		err = issueAuthToken(context)
	} else {
		if required {
			err = errors.New("Invalid Username/Password submitted")
		}
	}
	return
}
