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
	admin, _ := models.GetUserByUsernameAndDatabase(db, config.Config.Authentication.AdminUsername)
	if admin == nil {
		models.InsertUserWithUsernameAndDatabase(db, config.Config.Authentication.AdminUsername, config.Config.Authentication.AdminPassword, "UTC", true)
	}
}

func authenticate(context *util.Context, authType AuthenticationType) (err error) {
	switch authType {

	case NoAuthentication:
		return

	case DeviceAuthentication:
		err = authenticateCredentials(context, true)

	case PasswordAuthentication:
		err = authenticatePassword(context, true)

	case TokenAuthentication:
		err = authenticateToken(context, true)

	case AnyAuthentication:
		err = authenticatePassword(context, false)
		if err == nil {
			err = authenticateToken(context, false)
			if err == nil {
				err = authenticateCredentials(context, true)
			}
		}
	}
	return
}
