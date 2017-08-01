package models

import (
	"bloodtales/util"
	"fmt"
)

const TutorialCollectionName = "tutorial"

type Tutorial struct {
	Name     string `bson:"nm" json:"name"`
	Complete bool   `bson:"cp" json:"complete"`
}

func (tutorial *Tutorial) initialize() {
	tutorial.Name = ""
	tutorial.Complete = false
}

func UpdateTutorial(context *util.Context, player *Player, name string, complete bool) (err error) {
	// Init Tutorial Data
	tutorial := Tutorial{
		Name:     name,
		Complete: complete,
	}

	dataExist := false
	for i := 0; i < len(player.Tutorial); i++ {
		if player.Tutorial[i].Name == tutorial.Name {
			player.Tutorial[i].Complete = tutorial.Complete
			dataExist = true
			break
		}
	}
	if !dataExist {
		player.Tutorial = append(player.Tutorial, tutorial)
	}

	err = player.Save(context)
	if err != nil {
		fmt.Println("[TUTORIAL] Save error")
		return
	}

	return
}
