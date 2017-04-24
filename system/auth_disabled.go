// +build noauth

package system

import (
	"bloodtales/models"
	"bloodtales/log"
)

var (
	debugUser *models.User = nil
)

func (application *Application) initializeAuthentication() {
	application.initializeToken()

	// find debug user instead of authenticating
	debugUsername := application.Env.GetString("DEBUG_USER", "")
	if debugUsername != "" {
		debugUser, _ = models.GetUserByUsername(application.db, debugUsername)
		
		if debugUser != nil {
			log.Warningf("DEBUG - Build has disabled authentication, using debug user: %v", debugUsername)
			return
		}
	}

	log.Warning("DEBUG - Build has disabled authentication, no debug user found")
}

func (context *Context) authenticate(authType AuthenticationType) error {
	context.User = debugUser
	return nil
}
