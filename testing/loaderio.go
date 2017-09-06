package testing

import (
	"bloodtales/util"
	"bloodtales/system"
	"bloodtales/log"
)

func init() {
	// loader.io verification token
	verificationToken := util.Env.GetString("LOADERIO_TOKEN", "")

	if verificationToken != "" {
		// route to return token
		system.App.HandleAPI("/" + verificationToken + "/", system.NoAuthentication, func(context *util.Context) {
			context.Write([]byte(verificationToken))
		})
	}
}
