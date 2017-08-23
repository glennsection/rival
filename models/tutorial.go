package models

import (
	"bloodtales/util"
	"fmt"
)

const TutorialCollectionName = "tutorial"

type Tutorial struct {
	Name     string `bson:"nm" json:"name"`
	Complete bool   `bson:"cp" json:"complete"`
	Page     int    `bson:"pg" json:"page"`
	Progress int    `bson:"ps" json:"progress"`
}

func (tutorial *Tutorial) initialize() {
	tutorial.Name = ""
	tutorial.Complete = false
	tutorial.Page = 0
	tutorial.Progress = 0
}

func UpdateTutorial(context *util.Context, player *Player, name string, complete bool, page int, progress int) (err error) {
	// Init Tutorial Data
	tutorial := Tutorial{
		Name:     name,
		Complete: complete,
		Page:     page,
		Progress: progress,
	}

	dataExist := false
	for i := 0; i < len(player.Tutorial); i++ {
		if player.Tutorial[i].Name == tutorial.Name {
			player.Tutorial[i].Complete = tutorial.Complete
			player.Tutorial[i].Page = tutorial.Page
			player.Tutorial[i].Progress = tutorial.Progress
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
