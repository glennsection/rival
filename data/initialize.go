package data

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"bloodtales/util"	
)

// root data directory
var url = "https://s3-us-west-1.amazonaws.com/bloodtalesdev/btMain"
var rootDirectory string = "%v/server/GameData/%v"

// initialize data system
func init() {
	// create data table
	dataIdMap = map[DataId]string {}

	// load all data files
	// ------------------------------------------
	loadDataFile("ExcelConverted/PlayerLevelProgression.json", LoadPlayerLevelProgression)
	loadDataFile("ExcelConverted/Cards.json", LoadCards)
	loadDataFile("ExcelConverted/CommonCardLeveling.json", LoadCommonCardProgression)
	loadDataFile("ExcelConverted/RareCardLeveling.json", LoadRareCardProgression)
	loadDataFile("ExcelConverted/EpicCardLeveling.json", LoadEpicCardProgression)
	loadDataFile("ExcelConverted/LegendaryCardLeveling.json", LoadLegendaryCardProgression)
	loadDataFile("ExcelConverted/Tomes.json", LoadTomes)
	loadDataFile("ExcelConverted/TomeOrder.json", LoadTomeOrder)
	loadDataFile("ExcelConverted/PvPRanking.json", LoadRanks)
	loadDataFile("ExcelConverted/Rewards.json", LoadRewardData)
	loadDataFile("ExcelConverted/Store.json", LoadStore)
	loadDataFile("ExcelConverted/CardPurchaseCosts.json", LoadCardPurchaseCosts)
	loadDataFile("ExcelConverted/Rarity.json", LoadRarityData)
	loadDataFile("Definitions/QuestTypes.txt", LoadQuestData)
	// ------------------------------------------

	// template funcs
	util.AddTemplateFunc("toDataName", ToDataName)
}

// load a particular file into a container
func loadDataFile(fileName string, processor func([]byte)) {
	// read file
	pathUrl := fmt.Sprintf(rootDirectory, url, fileName)
	print(pathUrl)

	rawUrl, errUrl := http.Get(pathUrl)
	defer rawUrl.Body.Close()
	util.Must(errUrl)

	body, errUrl := ioutil.ReadAll(rawUrl.Body)
	util.Must(errUrl)	

	// process
	processor(body)
}