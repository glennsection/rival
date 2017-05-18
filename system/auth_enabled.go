// +build !noauth

package system

import (
	"bloodtales/config"
	"bloodtales/util"
	"bloodtales/models"
)

func init() {
	// get database connection
	db := util.GetDatabaseConnection()
	defer db.Session.Close()

	// init admin user
	admin, _ := models.GetUserByUsername(db, config.Config.Authentication.AdminUsername)
	if admin == nil {
		models.InsertUserWithUsername(db, config.Config.Authentication.AdminUsername, config.Config.Authentication.AdminPassword, true)
	}
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
