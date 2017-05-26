package system

import (
	"time"
	"fmt"
	"errors"

	"github.com/dgrijalva/jwt-go"

	"bloodtales/config"
	"bloodtales/util"
	"bloodtales/models"
)

var (
	authenticationSecret []byte
)

func init() {
	// get secret from config
	authenticationSecret = []byte(config.Config.Authentication.TokenSecret)
}

func authenticateToken(context *util.Context, required bool) (err error) {
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
					var user *models.User
					user, err = models.GetUserByUsername(context.DB, username)

					if user == nil {
						err = errors.New(fmt.Sprintf("Failed to find user indicated by authentication token: %v (%v)", username, err))
					}

					SetUser(context, user)
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

func AppendAuthToken(context *util.Context) (err error) {
	user := GetUser(context)
	if user == nil {
		err = errors.New("No User set for context to apply auth token")
		return
	}

	// create auth token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
		"username": user.Username,
		"exp": time.Now().Add(time.Second * config.Config.Authentication.TokenExpiration).Unix(),
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

func ClearAuthToken(context *util.Context) {
	// clear auth token from session
	context.Session.Set("token", "")
	context.Session.Save()
}