// +build !noauth

package system

import (
	"time"
	"fmt"

	"github.com/dgrijalva/jwt-go"

	"bloodtales/models"
)

const authTokenSecret string = "5UP3R-53CR3T-T0K3N" // TODO - move this to a config

type AuthenticationType int

const (
	NoAuthentication AuthenticationType = iota
	AnyAuthentication
	PasswordAuthentication
	TokenAuthentication
)

func (application *Application) initializeAuthentication() {
}

func (application *Application) authenticate(session *Session, authType AuthenticationType) (err error) {
	// check auth type
	if authType == NoAuthentication {
		return
	}
	allowToken := (authType == AnyAuthentication || authType == TokenAuthentication)
	allowPassword := (authType == AnyAuthentication || authType == PasswordAuthentication)

	// check for token paremeter
	unparsedToken := session.GetParameter("token", "")

	if unparsedToken == "" {
		if allowPassword {
			// parse login parameters
			username, password := session.GetRequiredParameter("username"), session.GetRequiredParameter("password")

			// authenticate user
			session.User, err = models.LoginUser(session.Application.DB, username, password)
			if session.User == nil {
				panic(fmt.Sprintf("Invalid authentication information for username: %v (%v)", username, err))
			}

			// create auth token
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
				"username": username,
				"exp": time.Now().Add(time.Hour).Unix(),
			})

			// analytics tracking (TODO - integrate with session)
			//session.Track("Login", bson.M { "mood": "happy" })

			// sign and get the complete encoded token as string
			session.Token, err = token.SignedString([]byte(authTokenSecret))
		} else {
			panic("Invalid authentication method")
		}
	} else {
		if allowToken {
			// keep token in session
			session.Token = unparsedToken

			// parse token parameter
			var token *jwt.Token
			token, err = jwt.Parse(unparsedToken, func(token *jwt.Token) (interface{}, error) {
				// validate the alg is what you expect
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					panic(fmt.Sprintf("Unexpected auth signing method: %v", token.Header["alg"]))
				}

				//return myLookupKey(token.Header["kid"]), nil  // what is this?????
				return []byte(authTokenSecret), nil
			})

			// get user if valid token
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if username, ok := claims["username"].(string); ok {
						session.User, err = models.GetUserByUsername(application.DB, username)
						if session.User == nil {
							panic(fmt.Sprintf("Failed to find user indicated by authentication token: %v (%v)", username, err))
						}
					}
				}
			}
		} else {
			panic("Invalid authentication method")
		}
	}

	return
 }