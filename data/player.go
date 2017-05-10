package data

import (
	"encoding/json"
)

type LevelProgression struct {
	ID 				int 		`json:"id,string"`
	XpRequired 		int 		`json:"xpRequired,string"`
}

var playerLevelProgression []LevelProgression

//internal parsing data
type LevelRequirementsParsed struct {
	PlayerLevelProgression []LevelProgression
}

// data processor
func LoadPlayerLevelProgression(raw []byte) {
	// parse
	container := &LevelRequirementsParsed {}
	json.Unmarshal(raw, container)

	// enter into system data
	playerLevelProgression = container.PlayerLevelProgression
}

func GetAccountLevel(xp int) (level int) {
	for _, levelProgression := range playerLevelProgression {
		if xp >= levelProgression.XpRequired {
			level = levelProgression.ID
		} else {
			break
		}
	}

	return level
}

