package data

import(
	"encoding/json"
	"strconv"

	"bloodtales/util"
)

type TutorialReward struct {
	Name 			string 
	TomeID 			*DataId 			
	RewardID 		*DataId 
	RankPoints 		int 			
}

type TutorialRewardClient struct {
	Name 			string 			`json:"tutorialName"`
	TomeID 			string 			`json:"tomeId"`
	RewardID 		string 			`json:"rewardId"`
	RankPoints 		string 			`json:"rankPoints"`
}

var tutorialRewards map[string]*TutorialReward

type TutorialRewardsParsed struct {
	TutorialRewards []TutorialReward
}

// custom unmarshalling
func (tutorialReward *TutorialReward)UnmarshalJSON(raw []byte) error {
	// create client model
	client := &TutorialRewardClient {}
		
	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	tutorialReward.Name = client.Name

	if client.TomeID != "" {
		id := ToDataId(client.TomeID)
		tutorialReward.TomeID = &id
	} else {
		tutorialReward.TomeID = nil
	}

	if client.RewardID != "" {
		id := ToDataId(client.RewardID)
		tutorialReward.RewardID = &id
	} else {
		tutorialReward.RewardID = nil
	}

	if num, err := strconv.ParseInt(client.RankPoints, 10, 64); err == nil {
		tutorialReward.RankPoints = int(num)
	} else {
		tutorialReward.RankPoints = 0
	}

	return nil
}

// load data
func LoadTutorialRewards(raw []byte) {
	container := &TutorialRewardsParsed{}
	util.Must(json.Unmarshal(raw, container))

	tutorialRewards = map[string]*TutorialReward{}

	for i, _ := range container.TutorialRewards {
		tutorialRewards[container.TutorialRewards[i].Name] = &container.TutorialRewards[i]
	}

	DebugTutRewards()
}

func GetTutorialReward(name string) *TutorialReward {
	if reward, contains := tutorialRewards[name]; contains {
		return reward
	}

	return nil
}