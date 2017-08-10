package config

import (
	"time"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type LogLevel string
const (
	NoLogging LogLevel = "NoLogging"
	BriefLogging = "BriefLogging"
	FullLogging = "FullLogging"
)

type Configuration struct {
	Platform struct {
		Version           string           `json:"Version"`
	}
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
		OfflineTimeout    time.Duration    `json:"OfflineTimeout,int"`
	}                                      `json:"Sessions"`
	Resources struct {
		DataPath          string           `json:"DataPath"`
	}                                      `json:"Resources"`
	Logging struct {
		Requests          LogLevel         `json:"Requests"`
	}                                      `json:"Logging"`
}

var Config Configuration

func Load(path string, config *Configuration) (err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, config)
	return
}

func init() {
	// load config
	configPath := "./config.json"
	err := Load(configPath, &Config)
	if err != nil {
		panic(fmt.Sprintf("Config file (%s) failed to load: %v", configPath, err))
	}
}
