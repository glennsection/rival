package data

import (
	"encoding/json"

	"bloodtales/util"
)

type GameplayConfiguration struct {
	FreeTomeUnlockTime 			int  		`json:"freeTomeUnlockTime"` //seconds
	BattleTomeCooldown 			int 		`json:"battleTomeCooldown"` //seconds
	LegendaryCardCurrencyValue 	int 		`json:"legendaryCardCurrencyValue"`

	GuildMemberLimit			int			`json:"guildMemberLimit"`
}

var GameplayConfig GameplayConfiguration
var GameplayConfigJSON string

// data processor
func LoadGameplayConfig(raw []byte) {
	GameplayConfigJSON = string(raw)
	
	util.Must(json.Unmarshal(raw, &GameplayConfig))
}
