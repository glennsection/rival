package system

import (
	"fmt"
	"errors"

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

func SetUser(context *Context, user *models.User) {
	context.Params.Set("user", user)
}

func GetUser(context *Context) *models.User {
	if user, ok := context.Params.Get("user").(*models.User); ok {
		return user
	}
	return nil
}

// check if context is authenticated
func (context *Context) Authenticated() bool {
	return GetUser(context) != nil
}

// basic username/password auth
func (context *Context) authenticatePassword(required bool) (err error) {
	// parse login parameters
	username, password := context.Params.GetString("username", ""), context.Params.GetString("password", "")

	// authenticate user
	if username != "" && password != "" {
		var user *models.User
		user, err = models.LoginUser(context.DB, username, password)
		if user == nil {
			err = errors.New(fmt.Sprintf("Invalid authentication information for username: %v (%v)", username, err))
			return
		}

		// set user in context
		SetUser(context, user)

		err = context.AppendAuthToken()
	} else {
		if required {
			err = errors.New("Invalid Username/Password submitted")
		}
	}
	return
}
