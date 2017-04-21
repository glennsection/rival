// +build !noauth

package system

import (
	"time"
	"fmt"
	"errors"

	"github.com/dgrijalva/jwt-go"

	"bloodtales/models"
)

type AuthenticationType int

const (
	NoAuthentication AuthenticationType = iota
	AnyAuthentication
	PasswordAuthentication
	TokenAuthentication
)

var (
	authenticationSecret []byte
)

func (application *Application) initializeAuthentication() {
	// get secret from config
	authenticationSecret = []byte(application.Config.Authentication.Secret)
}

func (context *Context) authenticate(authType AuthenticationType) (err error) {
	switch authType {

	case NoAuthentication:
		return

	case PasswordAuthentication:
		err = context.authenticatePassword(true)

	case TokenAuthentication:
		err = context.authenticateToken(true)

	case AnyAuthentication:
		err = context.authenticatePassword(false)
		if err == nil {
			err = context.authenticateToken(true)
		}
	}
	return
}

func (context *Context) authenticatePassword(required bool) (err error) {
	// parse login parameters
	username, password := context.Params.GetString("username", ""), context.Params.GetString("password", "")

	// authenticate user
	if username != "" && password != "" {
		context.User, err = models.LoginUser(context.DB, username, password)
		if context.User == nil {
			err = errors.New(fmt.Sprintf("Invalid authentication information for username: %v (%v)", username, err))
			return
		}

		err = context.AppendToken()
	} else {
		if required {
			err = errors.New("Invalid Username/Password submitted")
		}
	}
	return
}

func (context *Context) authenticateToken(required bool) (err error) {
	// check for token first in URL parameters
	unparsedToken := context.Params.GetString("token", "")
	if unparsedToken == "" {
		// if not found, then check session
		unparsedToken, _ = context.Session.Values["token"].(string)
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

func (context *Context) Authenticated() bool {
	return context.User != nil
}

func (context *Context) AppendToken() (err error) {
	// create auth token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
		"username": context.User.Username,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	// sign and get the complete encoded token as string
	context.Token, err = token.SignedString(authenticationSecret)

	// store token in session
	if err == nil {
		context.Session.Values["token"] = context.Token
		context.Session.Save(context.Request, context.responseWriter)
	}
	return
}

func (context *Context) ClearAuthentication() {
	context.Session.Values["token"] = ""
	context.Session.Save(context.Request, context.responseWriter)
}