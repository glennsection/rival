package system

import (
	"os"
	"fmt"
	"net/http"
	"encoding/gob"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/heroku"
	gocontext "golang.org/x/net/context"

	"bloodtales/config"
	"bloodtales/util"
)

var (
	oauthConfig	  *oauth2.Config
	oauthStateToken  string
)

func init() {
	oauthConfig = &oauth2.Config {
		ClientID: config.Config.Authentication.OAuthID,
		ClientSecret: config.Config.Authentication.OAuthSecret,
		RedirectURL: fmt.Sprintf("%s/auth/callback", os.Getenv("APP_URL")),
		Scopes: []string {
			"identity",
			//"read",
			//"write",
			//"openid",
			// "https://www.googleapis.com/auth/userinfo.profile", //https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: heroku.Endpoint,
	}

	oauthStateToken = config.Config.Authentication.OAuthStateToken

	gob.Register(&oauth2.Token{})

	// init URL handlers
	url := oauthConfig.AuthCodeURL(oauthStateToken)
	App.Redirect("/auth", url, http.StatusFound)
	App.HandleAPI("/auth/callback", NoAuthentication, handleAuthCallback)
}

func handleAuthCallback(context *Context) {
	// parse parameters
	state := context.Params.GetRequiredString("state")
	code := context.Params.GetRequiredString("code")

	// check state value
	if state != oauthStateToken {
		panic(fmt.Sprintf("Invalid OAuth State Token: %s", state))
	}

	// exchange with default context
	ctx := gocontext.Background()
	token, err := oauthConfig.Exchange(ctx, code)
	util.Must(err)

	// session, err := context.application.cookies.Get(r, "heroku-oauth-example-go")
	// if err != nil {
	// 	panic(err)
	// }
	// session.Values["heroku-oauth-token"] = token
	context.Session.Set("oauth-token", token)
	util.Must(context.Session.Save())

	context.Redirect("/user", http.StatusFound) // TODO - where should it redirect for mobile?
}
