// +build !noauth

package system

import (
	"bloodtales/models"
)

func (application *Application) initializeAuthentication() {
	// init admin user
	admin, _ := models.GetUserByUsername(application.db, application.Config.Authentication.AdminUsername)
	if admin == nil {
		models.InsertUserWithUsername(application.db, application.Config.Authentication.AdminUsername, application.Config.Authentication.AdminPassword, true)
	}

	application.initializeToken()
	application.initializeOAuth()
}

func (context *Context) authenticate(authType AuthenticationType) (err error) {
	switch authType {

	case NoAuthentication:
		return

	case DeviceAuthentication:
		err = context.authenticateDevice(true)

	case PasswordAuthentication:
		err = context.authenticatePassword(true)

	case TokenAuthentication:
		err = context.authenticateToken(true)

	case AnyAuthentication:
		err = context.authenticatePassword(false)
		if err == nil {
			err = context.authenticateToken(false)
			if err == nil {
				err = context.authenticateDevice(true)
			}
		}
	}
	return
}
