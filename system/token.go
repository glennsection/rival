package system

import (
	"time"
	"fmt"
	"errors"

	"gopkg.in/mgo.v2/bson"
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
		if err != nil {
			return
		}

		// get user if valid token
		if !token.Valid {
			return errors.New(fmt.Sprintf("Invalid authentication token: %v", token))
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if id, ok := claims["id"].(string); ok {
				if bson.IsObjectIdHex(id) {
					var user *models.User
					user, err = models.GetUserById(context, bson.ObjectIdHex(id))

					if user == nil {
						err = errors.New(fmt.Sprintf("Failed to find user indicated by authentication token. ID: %v, Error: %v", id, err))
					} else {
						SetUser(context, user)
						return
					}
				} else {
					err = errors.New(fmt.Sprintf("Invalid ID claim from authentication token: %v", id))
				}
			} else {
				err = errors.New(fmt.Sprintf("Failed to retrieve ID claim from authentication token: %v", token))
			}
		} else {
			err = errors.New(fmt.Sprintf("Failed to retrieve claims from authentication token: %v", token))
		}
	} else {
		if required {
			err = errors.New("Unauthorized user")
		}
	}
	return
}

func issueAuthToken(context *util.Context) (err error) {
	user := GetUser(context)
	if user == nil {
		err = errors.New("No User set for context to apply auth token")
		return
	}

	// create auth token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
		"id": user.ID.Hex(),
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