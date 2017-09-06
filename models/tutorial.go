package models

import (
	"fmt"

	"bloodtales/util"
	"bloodtales/data"
)

const TutorialCollectionName = "tutorial"

type Tutorial struct {
	Name     string `bson:"nm" json:"name"`
	Complete bool   `bson:"cp" json:"complete"`
	Page     int    `bson:"pg" json:"page"`
	Progress int    `bson:"ps" json:"progress"`
	Rewarded bool 	`bson:"rd"` //client doesn't need to know about this
}

func (tutorial *Tutorial) initialize() {
	tutorial.Name = ""
	tutorial.Complete = false
	tutorial.Page = 0
	tutorial.Progress = 0
	tutorial.Rewarded = false
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
		tutorial.Rewarded = false
		player.Tutorial = append(player.Tutorial, tutorial)
	}

	err = player.Save(context)
	if err != nil {
		fmt.Println("[TUTORIAL] Save error")
		return
	}

	return
}

func (player *Player)ClaimTutorialReward(context *util.Context, name string) (tome *Tome, reward *Reward, err error) {
	var tutorial *Tutorial

	for i := range player.Tutorial {
		if player.Tutorial[i].Name == name {
			if player.Tutorial[i].Rewarded {
				return
			}

			player.Tutorial[i].Rewarded = true
			tutorial = &player.Tutorial[i]
			break
		}
	}

	if tutorial == nil {
		tutorial = &Tutorial{
			Name: name,
			Complete: false,
			Page: 0,
			Progress: 0,
			Rewarded: true,
		}
		player.Tutorial = append(player.Tutorial, *tutorial)
	}

	tutorialReward := data.GetTutorialReward(name)
	if tutorialReward == nil {
		return
	}

	if tutorialReward.TomeID != nil {
		fmt.Println(fmt.Sprintf("Tutorial %s is rewarding %s", name, data.ToDataName(*tutorialReward.TomeID)))

		_, tome = player.GetEmptyTomeSlot()
		
		if tome != nil {
			tome.DataID = *tutorialReward.TomeID
			tome.State = TomeLocked
			tome.UnlockTime = 0
			tome.League = data.GetLeague(data.GetRank(player.RankPoints).Level)
		}
	}

	if tutorialReward.RewardID != nil {
		reward = player.GetReward(*tutorialReward.RewardID, data.LeagueZero, 1)
		player.AddRewards(reward, nil)
	}

	player.RankPoints += tutorialReward.RankPoints

	err = player.Save(context)

	return
}