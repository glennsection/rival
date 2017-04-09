// +build noauth

package system

import (
	"log"
	
	"bloodtales/models"
)

type AuthenticationType int

const (
	NoAuthentication AuthenticationType = iota
	AnyAuthentication
	PasswordAuthentication
	TokenAuthentication
)

func (application *Application) initializeAuthentication() {
	log.Printf("DEBUG - Build has authentication disabled")
}

func (application *Application) authenticate(session *Session, authType AuthenticationType) (err error) {
	// find debug user instead of authenticating
	debugUser := application.GetEnv("DEBUG_USER", "")
	if debugUser != "" {
		session.User, err = models.GetUserByUsername(application.DB, debugUser)
		
		if session.User != nil {
			log.Printf("DEBUG - Authentication disabled, using debug user: %v", debugUser)
			return
		}
	}

	log.Printf("DEBUG - Authentication disabled, no debug user found")
	return
 }