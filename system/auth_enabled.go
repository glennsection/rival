// +build !noauth

package system

func (application *Application) initializeAuthentication() {
	application.initializeToken()
	application.initializeOAuth()
}

func (context *Context) authenticate(authType AuthenticationType) (err error) {
	switch authType {

	case NoAuthentication:
		return

	case DeviceAuthentication:
		err = context.authenticateDevice(true)

	case PasswordAuthentication:
		err = context.authenticatePassword(true)

	case TokenAuthentication:
		err = context.authenticateToken(true)

	case AnyAuthentication:
		err = context.authenticatePassword(false)
		if err == nil {
			err = context.authenticateToken(false)
			if err == nil {
				err = context.authenticateDevice(true)
			}
		}
	}
	return
}
