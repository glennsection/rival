package data

import (
	"encoding/json"

	"bloodtales/util"
)

type RarityData struct {
	ID 		 					string 		`json:"id"`
	MaxLevel 					int 		`json:"maxLevel,string"`
	DamageLevelMultiplier		float64		`json:"damageLevelMultiplier,string"`
	HealthLevelMultiplier 		float64 	`json:"healthLevelMultiplier,string"`
	CraftingCost 				int 		`json:"craftingCost,string"`
	CraftingXp 					int 		`json:"craftingXp,string"`
	CraftingXpNeeded			int 		`json:"craftingXpNeeded,string"`
}

var rarityData map[DataId]*RarityData

type RarityDataParsed struct {
	Rarity []RarityData
}

func LoadRarityData(raw []byte) {
	// parse
	container := &RarityDataParsed {}
	json.Unmarshal(raw, container)

	// enter into system data
	rarityData = map[DataId]*RarityData {}
	for i, rarity := range container.Rarity {
		// map name to ID
		id, err := mapDataName(rarity.ID)
		util.Must(err)

		// insert into table
		rarityData[id] = &container.Rarity[i]
	}
}

func GetMaxLevel(rarity string) int {
	return rarityData[ToDataId(rarity)].MaxLevel
}

func GetCraftingCost(rarity string) int {
	return rarityData[ToDataId(rarity)].CraftingCost
}

func GetCraftingXp(rarity string) int {
	return rarityData[ToDataId(rarity)].CraftingXp
}

func GetCraftingXpNeeded(rarity string) int {
	return rarityData[ToDataId(rarity)].CraftingXpNeeded
}