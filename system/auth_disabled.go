// +build noauth

package system

import (
	"bloodtales/util"
	"bloodtales/models"
	"bloodtales/log"
)

var (
	debugUser *models.User = nil
)

func init() {
	// get database connection
	db := util.GetDatabaseConnection()
	defer db.Session.Close()

	// find debug user instead of authenticating
	debugUUID := util.Env.GetString("DEBUG_USER", "")
	if debugUUID != "" {
		debugUser, _ = models.GetUserByUUID(db, debugUUID)
		
		if debugUser != nil {
			log.Warningf("DEBUG - Build has disabled authentication, using debug user: %v", debugUUID)
			return
		}
	}

	log.Warning("DEBUG - Build has disabled authentication, no debug user found")
}

func authenticate(context *util.Context, authType AuthenticationType) error {
	SetUser(context, debugUser)
	return nil
}
