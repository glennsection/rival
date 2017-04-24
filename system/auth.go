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
	PasswordAuthentication
	TokenAuthentication
)

// check if context is authenticated
func (context *Context) Authenticated() bool {
	return context.User != nil
}

// basic username/password auth
func (context *Context) authenticatePassword(required bool) (err error) {
	// parse login parameters
	username, password := context.Params.GetString("username", ""), context.Params.GetString("password", "")

	// authenticate user
	if username != "" && password != "" {
		context.User, err = models.LoginUser(context.DB, username, password)
		if context.User == nil {
			err = errors.New(fmt.Sprintf("Invalid authentication information for username: %v (%v)", username, err))
			return
		}

		err = context.AppendAuthToken()
	} else {
		if required {
			err = errors.New("Invalid Username/Password submitted")
		}
	}
	return
}
