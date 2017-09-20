package config

import (
	"time"
	"fmt"
	"os"
	"strconv"
	"io/ioutil"
	"encoding/json"
)

type LogLevel string
const (
	NoLogging LogLevel = "NoLogging"
	BriefLogging = "BriefLogging"
	FullLogging = "FullLogging"
)

type LoggingConfiguration struct {
		Requests            LogLevel
}

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
		Production          LoggingConfiguration
		Development         LoggingConfiguration
	}
	Matches struct {
		MatchTicketExpire   int
		MatchResultExpire   int
		MaxMMRDeltas        []int
	}
}

type Environment struct {
	Name                    string
	Development             bool
}

var Config Configuration
var Env Environment

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

	// initialize environment
	developmentString := os.Getenv("DEVELOPMENT")
	development, err := strconv.ParseBool(developmentString)
	if err == nil {
		Env.Development = development
	}
	if Env.Development {
		Env.Name = "Development"
	} else {
		Env.Name = "Production"
	}
}

func (config *Configuration) GetLogging() (logging *LoggingConfiguration) {
	if Env.Development {
		return &config.Logging.Development;
	} else {
		return &config.Logging.Production;
	}
}