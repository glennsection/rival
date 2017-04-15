// +build noauth

package system

import (
	"time"
	
	"github.com/dgrijalva/jwt-go"

	"bloodtales/models"
	"bloodtales/log"
)

const authTokenSecret string = "5UP3R-53CR3T-T0K3N" // TODO - move this to a config

type AuthenticationType int

const (
	NoAuthentication AuthenticationType = iota
	AnyAuthentication
	PasswordAuthentication
	TokenAuthentication
)

var (
	debugUser *models.User = nil
)

func (application *Application) initializeAuthentication() {
	// find debug user instead of authenticating
	debugUsername := application.GetEnv("DEBUG_USER", "")
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

func (context *Context) AppendToken() (err error) {
	// create auth token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
		"username": context.User.Username,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	// analytics tracking (TODO - integrate with context)
	//context.Track("Login", bson.M { "mood": "happy" })

	// sign and get the complete encoded token as string
	context.Token, err = token.SignedString([]byte(authTokenSecret))
	return
}