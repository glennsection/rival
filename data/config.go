package data

import (
	"bloodtales/util"
	"encoding/json"
)

type Configuration struct {
	General  				GeneralConfiguration 	`json:"general"`
}

type GeneralConfiguration struct {
	FreeTomeUnlockTime 		int64  					`json:"freeTomeUnlockTime"`
}

var config *Configuration
var configFile *string

// data processor
func LoadConfig(raw []byte) {
	// parse
	util.Must(json.Unmarshal(raw, config))

	*configFile = string(raw)
}

func Config() Configuration {
	return *config
}

func GetConfigFile() string {
	return *configFile
}
