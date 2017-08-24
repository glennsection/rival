package data

import (
	"encoding/json"

	"bloodtales/util"
)

type QuestPeriod int
const (
	QuestPeriodDaily QuestPeriod = iota
	QuestPeriodWeekly
	QuestPeriodEvent
)

type QuestType int
const (
	QuestTypeBattle QuestType = iota
)

type QuestPhaseData struct {
	Objective 			int						`json:"objective"`
	RewardID 			DataId
}

type QuestData struct {
	Name				string 					`json:"id"`
	Period 				QuestPeriod
	Type 				QuestType
	Disposable 			bool 					`json:"disposable"`		
	Time 				int64 					`json:"time"`
	PercentChance		float32 				`json:"percentChance"`
	Phases 				[]QuestPhaseData

	Properties 			map[string]interface{}	`json:"properties"`
}

type QuestPhaseDataClientAlias QuestPhaseData
type QuestPhaseDataClient struct {
	RewardID 			string 					`json:"rewardId"`

	*QuestPhaseDataClientAlias
}

type QuestDataClientAlias QuestData
type QuestDataClient struct {
	Period 				string 					`json:"period"`
	Type 				string 					`json:"type"`
	Phases 				[]QuestPhaseDataClient	`json:"phases"`

	*QuestDataClientAlias
}

// data map
var quests map[DataId]*QuestData

// supported periods per quest slot
var supportedQuestPeriods [][]QuestPeriod

func GetQuestData(id DataId) *QuestData {
	return quests[id]
}

// implement Data interface
func (data *QuestData) GetDataName() string {
	return data.Name
}

// internal parsing data (TODO - ideally we'd just remove this top-layer from the JSON files)
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
	switch client.Period {

	case "Daily":
		quest.Period = QuestPeriodDaily

	case "Weekly":
		quest.Period = QuestPeriodWeekly

	default:
		quest.Period = QuestPeriodEvent

	}

	// phases
	quest.Phases = make([]QuestPhaseData, len(client.Phases))
	for i, _ := range client.Phases {
		quest.Phases[i].RewardID = ToDataId(client.Phases[i].RewardID)
		quest.Phases[i].Objective = client.Phases[i].Objective
	}

	// assign objectives and properties
	quest.Properties = client.Properties

	return nil
}

// data processor
func LoadQuestData(raw []byte) {
	// parse and enter into system data
	container := &QuestDataParsed{}
	util.Must(json.Unmarshal(raw, container))

	quests = map[DataId]*QuestData {}
	for i, quest := range container.QuestTypes {
		id, err := mapDataName(quest.Name)
		util.Must(err)

		// insert into table
		quests[id] = &container.QuestTypes[i]
	}

	// supported quest periods
	supportedQuestPeriods = make([][]QuestPeriod, 3, 3)
	supportedQuestPeriods[0] = []QuestPeriod { QuestPeriodDaily }
	supportedQuestPeriods[1] = []QuestPeriod { QuestPeriodDaily }
	supportedQuestPeriods[2] = []QuestPeriod { QuestPeriodWeekly }
}

func (questData *QuestData) IsSupported(index int) bool {
	if index >= 0 && index < len(supportedQuestPeriods) {
		for _, period := range supportedQuestPeriods[index] {
			if period == questData.Period {
				return true
			}
		}
	}
	return false
}

func GetRandomQuestData(condition func(DataId, *QuestData) bool) (dataId DataId, questData *QuestData) {
	for id, quest := range quests {
		if condition(id, quest) { // condition should be defined {return true} for any quest to be returned
			dataId = id
			questData = quest
			return // we can break after the first successful iteration because items in golang maps are accessed in randomized order
		}
	}

	return
}