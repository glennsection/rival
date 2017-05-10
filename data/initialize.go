package data

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// root data directory
var rootDirectory string = "./resources/data/%v.json"

// initialize data system
func Load() {
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
	// ------------------------------------------
}

// load a particular file into a container
func loadDataFile(fileName string, processor func([]byte)) {
	// read file
	path, err := filepath.Abs(fmt.Sprintf(rootDirectory, fileName))
	if err != nil {
		panic(err)
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// process
	processor(raw)
}