package system

import (
	"github.com/gorilla/sessions"
)

type Session struct {
	Stream

	// internal
	session          *sessions.Session
	context          *Context
}

type SessionStreamSource struct {
	// internal
	session          *sessions.Session
}

var (
	// internal
	cookieStore		 *sessions.CookieStore
)

func (source SessionStreamSource) Has(name string) bool {
	_, ok := source.session.Values[name]
	return ok
}

func (source SessionStreamSource) Set(name string, value interface{}) {
	source.session.Values[name] = value
}

func (source SessionStreamSource) Get(name string) interface{} {
	if value, ok := source.session.Values[name]; ok {
		return value
	}
	return ""
}

func (application *Application) initializeSessions() {
	cookieSecret := application.Config.Sessions.CookieSecret
	cookieStore = sessions.NewCookieStore([]byte(cookieSecret))

	//cookie.SetSerializer(securecookie.JSONEncoder{})

	//cookieStore.MaxAge(60 * 60 * 8) // 8 hour expiration
	//cookieStore.Options.Secure = true // secure for OAuth
}

func (context *Context) getSession() (session *Session) {
	// get cookis session from store
	cookieSession, err := cookieStore.Get(context.Request, "session")
	if err != nil {
		panic(err)
	}

	// stream source
	source := SessionStreamSource {
		session: cookieSession,
	}

	// create abstracted session
	session = &Session {
		Stream: Stream {
			source: source,
		},
		session: cookieSession,
		context: context,
	}
	return
}

func (session *Session) Save() error {
	return session.session.Save(session.context.Request, session.context.responseWriter)
}
