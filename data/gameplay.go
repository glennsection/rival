package data

import (
	"encoding/json"

	"bloodtales/util"
)

type GameplayConfiguration struct {
	Arenas                      []string    `json:"arenas"`

	FreeTomeUnlockTime 			int  		`json:"freeTomeUnlockTime"` //seconds
	BattleTomeCooldown 			int 		`json:"battleTomeCooldown"` //seconds

	LegendaryCardCurrencyValue 	int			`json:"legendaryCardCurrencyValue"`

	PeriodicOfferCooldown 		int 		`json:"periodicOfferCooldown"` //days

	GuildMemberLimit			int			`json:"guildMemberLimit"`
	GuildCreationCost           int			`json:"guildCreateCost"`
	MaxGuildNameLength          int			`json:"maxGuildNameLength"`
	MinGuildNameLength			int			`json:"minGuildNameLength"`
	MaxGuildDescriptionLength   int			`json:"maxGuildDescriptionLength"`

	MinUsernameLength 			int 		`json:"minUsernameLength"`
	MaxUsernameLength 			int 		`json:"maxUsernameLength"`

	BatchLimit 					int			`json:"batchLimit"`
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

func GetRandomArena() string {
	index := util.RandomIntn(len(GameplayConfig.Arenas))
	return GameplayConfig.Arenas[index]
}