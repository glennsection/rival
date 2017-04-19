// +build !noauth

package system

import (
	"time"
	"fmt"
	"errors"

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

func (context *Context) authenticate(authType AuthenticationType) (err error) {
	// check auth type
	if authType == NoAuthentication {
		return
	}
	allowToken := (authType == AnyAuthentication || authType == TokenAuthentication)
	allowPassword := (authType == AnyAuthentication || authType == PasswordAuthentication)

	// check for token paremeter
	unparsedToken := context.Params.GetString("token", "")

	if unparsedToken == "" {
		if allowPassword {
			// parse login parameters
			username, password := context.Params.GetRequiredString("username"), context.Params.GetRequiredString("password")

			// authenticate user
			context.User, err = models.LoginUser(context.DB, username, password)
			if context.User == nil {
				err = errors.New(fmt.Sprintf("Invalid authentication information for username: %v (%v)", username, err))
				return
			}

			err = context.AppendToken()
		} else {
			err = errors.New("Invalid authentication method")
		}
	} else {
		if allowToken {
			// keep token in context
			context.Token = unparsedToken

			// parse token parameter
			var token *jwt.Token
			token, err = jwt.Parse(unparsedToken, func(token *jwt.Token) (interface{}, error) {
				// validate the alg is what you expect
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New(fmt.Sprintf("Unexpected auth signing method: %v", token.Header["alg"]))
				}

				return []byte(authTokenSecret), nil
			})

			// get user if valid token
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if username, ok := claims["username"].(string); ok {
						context.User, err = models.GetUserByUsername(context.DB, username)
						if context.User == nil {
							panic(fmt.Sprintf("Failed to find user indicated by authentication token: %v (%v)", username, err))
						}
					}
				}
			}
		} else {
			err = errors.New("Invalid authentication method")
		}
	}

	return
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