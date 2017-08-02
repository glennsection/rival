package data

import (
	"bloodtales/util"
	"encoding/json"
)

type Configuration struct {
	FreeTomeUnlockTime 			int  		`json:"freeTomeUnlockTime"` //seconds
	BattleTomeCooldown 			int 		`json:"battleTomeCooldown"` //seconds
	LegendaryCardCurrencyValue 	int 		`json:"legendaryCardCurrencyValue"`
}

var config *Configuration

type ConfigurationParsed struct {
	Config		 				Configuration
}

// data processor
func LoadConfig(raw []byte) {
	config = &Configuration{}
	util.Must(json.Unmarshal(raw, config))
}

func Config() Configuration {
	return *config
}
