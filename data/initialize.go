package data

import (
	"fmt"
	"strings"
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
	loadDataFile("GameData/ExcelConverted/TutorialRewards.json", LoadTutorialRewards)
	// ------------------------------------------

	// template funcs
	util.AddTemplateFunc("toDataName", ToDataName)
}

// load a particular file into a container
func loadDataFile(fileName string, processor func([]byte)) {
	// get file path
	dataPath := util.Env.GetRequiredString("DATA_URL")
	filePath := fmt.Sprintf("%s/%s", dataPath, fileName)

	var body []byte
	var err error

	if strings.HasPrefix(filePath, "file://") {
		// read file from local
		body, err = ioutil.ReadFile(filePath)
		util.Must(err)
	} else {
		// read file from URL
		var response *http.Response
		response, err = http.Get(filePath)
		defer response.Body.Close()
		util.Must(err)

		body, err = ioutil.ReadAll(response.Body)
		util.Must(err)
	}

	// process
	processor(body)
}
