package data

import (
	"encoding/json"

	"bloodtales/util"
)

type GameplayConfiguration struct {
	FreeTomeUnlockTime 			int  		`json:"freeTomeUnlockTime"` //seconds
	BattleTomeCooldown 			int 		`json:"battleTomeCooldown"` //seconds
	LegendaryCardCurrencyValue 	int
}

type GameplayConfigurationParsed struct {
	Config GameplayConfiguration		    `json:"gameplay"`
}

var GameplayConfig GameplayConfiguration
var GameplayConfigJSON string

// data processor
func LoadGameplayConfig(raw []byte) {
	GameplayConfigJSON = string(raw)
	
	parsed := &GameplayConfigurationParsed {}
	util.Must(json.Unmarshal(raw, &parsed))

	GameplayConfig = parsed.Config
}
