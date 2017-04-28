package system

import (
	"time"
	"fmt"
	"errors"

	"github.com/dgrijalva/jwt-go"

	"bloodtales/models"
)

var (
	authenticationSecret []byte
)

func (application *Application) initializeToken() {
	// get secret from config
	authenticationSecret = []byte(application.Config.Authentication.TokenSecret)
}

func (context *Context) authenticateToken(required bool) (err error) {
	// check for token first in URL parameters
	unparsedToken := context.Params.GetString("token", "")
	if unparsedToken == "" {
		// if not found, then check session
		unparsedToken = context.Session.GetString("token", "")
	}

	if unparsedToken != "" {
		// keep token in context
		context.Token = unparsedToken

		// parse token parameter
		var token *jwt.Token
		token, err = jwt.Parse(unparsedToken, func(token *jwt.Token) (interface{}, error) {
			// validate the alg is what you expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New(fmt.Sprintf("Unexpected auth signing method: %v", token.Header["alg"]))
			}

			return authenticationSecret, nil
		})

		// get user if valid token
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if username, ok := claims["username"].(string); ok {
					context.User, err = models.GetUserByUsername(context.DB, username)
					if context.User == nil {
						err = errors.New(fmt.Sprintf("Failed to find user indicated by authentication token: %v (%v)", username, err))
					}
				}
			}
		}
	} else {
		if required {
			err = errors.New("Unauthorized user")
		}
	}
	return
}

func (context *Context) AppendAuthToken() (err error) {
	// create auth token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
		"username": context.User.Username,
		"exp": time.Now().Add(time.Hour * context.Config.Authentication.TokenExpiration).Unix(),
	})

	// sign and get the complete encoded token as string
	context.Token, err = token.SignedString(authenticationSecret)

	// store auth token in session
	if err == nil {
		context.Session.Set("token", context.Token)
		context.Session.Save()
	}
	return
}

func (context *Context) ClearAuthToken() {
	// clear auth token from session
	context.Session.Set("token", "")
	context.Session.Save()
}