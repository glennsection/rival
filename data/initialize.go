package data

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"bloodtales/util"
)

// root data directory
var rootDirectory string = "./resources/data/%v.json"

// initialize data system
func init() {
	// create data table
	dataIdMap = map[DataId]string {}

	// load all data files
	// ------------------------------------------
	loadDataFile("PlayerLevelProgression", LoadPlayerLevelProgression)
	loadDataFile("Cards", LoadCards)
	loadDataFile("CommonCardLeveling", LoadCommonCardProgression)
	loadDataFile("RareCardLeveling", LoadRareCardProgression)
	loadDataFile("EpicCardLeveling", LoadEpicCardProgression)
	loadDataFile("LegendaryCardLeveling", LoadLegendaryCardProgression)
	loadDataFile("Tomes", LoadTomes)
	loadDataFile("PvPRanking", LoadRanks)
	loadDataFile("Store", LoadStore)
	loadDataFile("CardPurchaseCosts", LoadCardPurchaseCosts)
	loadDataFile("Rarity", LoadRarityData)
	// ------------------------------------------

	// template funcs
	util.AddTemplateFunc("toDataName", ToDataName)
}

// load a particular file into a container
func loadDataFile(fileName string, processor func([]byte)) {
	// read file
	path, err := filepath.Abs(fmt.Sprintf(rootDirectory, fileName))
	util.Must(err)

	raw, err := ioutil.ReadFile(path)
	util.Must(err)

	// process
	processor(raw)
}