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
		Version             string
	}
	Authentication struct {
		TokenSecret         string
		TokenExpiration     time.Duration    `json:"TokenExpiration,int"`

		DebugToken          string

		AdminUsername       string
		AdminPassword       string

		OAuthID             string
		OAuthSecret         string
		OAuthStateToken     string
	}
	Sessions struct {
		CookieSecret        string
		OfflineTimeout      time.Duration    `json:"OfflineTimeout,int"`
	}
	Logging struct {
		Requests            LogLevel
	}
	Matches struct {
		MatchTicketExpire   int
		MatchResultExpire   int
		MaxMMRDeltas        []int
	}
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
