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
	debugUsername := util.Env.GetString("DEBUG_USER", "")
	if debugUsername != "" {
		debugUser, _ = models.GetUserByUsername(db, debugUsername)
		
		if debugUser != nil {
			log.Warningf("DEBUG - Build has disabled authentication, using debug user: %v", debugUsername)
			return
		}
	}

	log.Warning("DEBUG - Build has disabled authentication, no debug user found")
}

func (context *Context) authenticate(authType AuthenticationType) error {
	SetUser(context, debugUser)
	return nil
}
