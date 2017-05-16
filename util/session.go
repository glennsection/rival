package util

import (
	"net/http"

	"github.com/gorilla/sessions"

	"bloodtales/config"
)

type Session struct {
	Stream

	// internal
	session          *sessions.Session
	request          *http.Request
	responseWriter   http.ResponseWriter
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

func init() {
	cookieSecret := config.Config.Sessions.CookieSecret
	cookieStore = sessions.NewCookieStore([]byte(cookieSecret))

	//cookie.SetSerializer(securecookie.JSONEncoder{})

	//cookieStore.MaxAge(60 * 60 * 8) // 8 hour expiration
	//cookieStore.Options.Secure = true // secure for OAuth
}

func GetSession(w http.ResponseWriter, r *http.Request) (session *Session) {
	// get cookis session from store
	cookieSession, err := cookieStore.Get(r, "session")
	Must(err)
	
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
		request: r,
		responseWriter: w,
	}
	return
}

func (session *Session) Save() error {
	return session.session.Save(session.request, session.responseWriter)
}
