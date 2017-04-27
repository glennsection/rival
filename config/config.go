package config

import (
	"time"
	"io/ioutil"
	"encoding/json"
)

type LogLevel string
const (
	NoLogging LogLevel = "NoLogging"
	BriefLogging = "BriefLogging"
	FullLogging = "FullLogging"
)

type Config struct {
	Authentication struct {
		TokenSecret       string           `json:"TokenSecret"`
		TokenExpiration   time.Duration    `json:"TokenExpiration,int"`

		DebugToken        string           `json:"DebugToken"`

		AdminUsername     string           `json:"AdminUsername"`
		AdminPassword     string           `json:"AdminPassword"`

		OAuthID           string           `json:"OAuthID"`
		OAuthSecret       string           `json:"OAuthSecret"`
		OAuthStateToken   string           `json:"OAuthStateToken"`
	}                                      `json:"Authentication"`
	Sessions struct {
		CookieSecret      string           `json:"CookieSecret"`
	}                                      `json:"Sessions"`
	Logging struct {
		Requests          LogLevel         `json:"Requests"`
	}                                      `json:"Logging"`
}

func Load(path string, config *Config) (err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, config)
	return
}