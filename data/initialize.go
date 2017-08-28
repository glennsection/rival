package data

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"bloodtales/util"
)

// initialize data system
func init() {
	// create data table
	dataIdMap = map[DataId]string{}

	// load all data files
	// ------------------------------------------
	loadDataFile("Default Gameplay Configuration.txt", LoadGameplayConfig)
	loadDataFile("GameData/ExcelConverted/PlayerLevelProgression.json", LoadPlayerLevelProgression)
	loadDataFile("GameData/ExcelConverted/Cards.json", LoadCards)
	loadDataFile("GameData/ExcelConverted/CommonCardLeveling.json", LoadCommonCardProgression)
	loadDataFile("GameData/ExcelConverted/RareCardLeveling.json", LoadRareCardProgression)
	loadDataFile("GameData/ExcelConverted/EpicCardLeveling.json", LoadEpicCardProgression)
	loadDataFile("GameData/ExcelConverted/LegendaryCardLeveling.json", LoadLegendaryCardProgression)
	loadDataFile("GameData/ExcelConverted/Tomes.json", LoadTomes)
	loadDataFile("GameData/ExcelConverted/TomeOrder.json", LoadTomeOrder)
	loadDataFile("GameData/ExcelConverted/PvPRanking.json", LoadRanks)
	loadDataFile("GameData/ExcelConverted/PvPLeagues.json", LoadLeagues)
	loadDataFile("GameData/ExcelConverted/Rewards.json", LoadRewardData)
	loadDataFile("GameData/ExcelConverted/Store.json", LoadStore)
	loadDataFile("GameData/ExcelConverted/PeriodicOfferTable.json", LoadPeriodicOfferTable)
	loadDataFile("GameData/ExcelConverted/Rarity.json", LoadRarityData)
	loadDataFile("GameData/Definitions/QuestTypes.txt", LoadQuestData)
	// ------------------------------------------

	// template funcs
	util.AddTemplateFunc("toDataName", ToDataName)
}

// load a particular file into a container
func loadDataFile(fileName string, processor func([]byte)) {
	// read file
	dataPath := util.Env.GetRequiredString("DATA_URL")
	pathUrl := fmt.Sprintf("%s/%s", dataPath, fileName)

	rawUrl, errUrl := http.Get(pathUrl)
	defer rawUrl.Body.Close()
	util.Must(errUrl)

	body, errUrl := ioutil.ReadAll(rawUrl.Body)
	util.Must(errUrl)

	// process
	processor(body)
}
