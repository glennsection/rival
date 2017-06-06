package data

import (
	"encoding/json"

	"bloodtales/util"
)

const MinutesTillWeeklyQuestExpires = 10080
const MinutesTillDailyQuestExpires = 1440
const QuestSlotCooldownTime = 5

type QuestType int
const (
	QuestType_Daily QuestType = iota
	QuestType_Weekly
	QuestType_Event
)

type QuestLogicType int
const (
	QuestLogicType_Battle QuestLogicType = iota
)

type QuestData struct { // must be embedded in all structs implementing QuestData
	ID						string 				`json:"id"`
	Description 			string 				`json:"description"`
	LogicType 				QuestLogicType 		
	Type 					QuestType 			
	RewardID 				DataId 				
	Disposable 				bool 				`json:"disposable"`		
	Time 					int64 				`json:"time"`
	PercentChance			float32 			`json:"percentChance"`
	Objectives 				map[string]interface{}	
}

type QuestDataClientAlias QuestData
type QuestDataClient struct {
	// Quests
	LogicType 				string 				`json:"questLogicType"`
	Type 					string 				`json:"type"`
	RewardID 				string 				`json:"rewardID"`

	// Victory Quests
	VictoryCount 			int 				`json:"victoryCount"`
	RequiresVictory 		bool 				`json:"requireVictory"`
	WinAsLeader 			bool 				`json:"asLeader"`
	UseRandomCard 			bool 				`json:"useRandomCard"`
	CardID 					string 				`json:"cardID"`

	*QuestDataClientAlias
}

var quests map[DataId]QuestData

type QuestDataParsed struct {
	QuestTypes []QuestData
}

// custom unmarshalling
func (quest *QuestData) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &QuestDataClient {
		QuestDataClientAlias: (*QuestDataClientAlias)(quest),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// quest type
	switch client.Type {
	case "Daily":
		quest.Type = QuestType_Daily
	case "Weekly":
		quest.Type = QuestType_Weekly
	default:
		quest.Type = QuestType_Event
	}

	// reward id
	quest.RewardID = ToDataId(client.RewardID)

	// assign objectives
	quest.Objectives = map[string]interface{}{}
	switch client.LogicType {

	case "Battle":
		quest.LogicType = QuestLogicType_Battle
		quest.Objectives["completionCondition"] = client.VictoryCount
		quest.Objectives["requiresVictory"] = client.RequiresVictory
		quest.Objectives["asLeader"] = client.WinAsLeader
		quest.Objectives["useRandomCard"] = client.UseRandomCard
		quest.Objectives["cardId"] = client.CardID
	default:
	}

	return nil
}

// data processor
func LoadQuestData(raw []byte) {
	// parse and enter into system data
	container := &QuestDataParsed{}
	util.Must(json.Unmarshal(raw, container))

	quests = map[DataId]QuestData {}
	for _,quest := range container.QuestTypes {
		id, err := mapDataName(quest.ID)
		util.Must(err)

		// insert into table
		quests[id] = quest
	}
}

func GetQuestData(id DataId) QuestData {
	return quests[id]
}

func GetRandomQuestData() (dataId DataId, questData QuestData) {
	for id,quest := range quests {
		dataId = id
		questData = quest
		break //we can break after the first iteration because items in golang maps are accessed in randomized order
	}

	return dataId, questData
}