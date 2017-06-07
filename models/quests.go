package models

import (
	"time"
	"encoding/json"
	//"fmt"
	
	"bloodtales/data"
	"bloodtales/util"
)

type QuestState int
const (
	QuestState_Ready QuestState = iota
	QuestState_InProgress
	QuestState_Collect
	QuestState_Cooldown
)

type QuestSlot struct {
	QuestInstance	Quest 					`bson:"qi"`
	State 			QuestState 				`bson:"qs"`
	UnlockTime 		int64 					`bson:"ut"`
	ExpireTime 		int64 					`bson:"et"`
	SupportedTypes	[]data.QuestType
}

type QuestSlotClient struct {
	QuestInstance 	string 					`json:"questInstance"`
	State 			string 					`json:"questSlotState"`
	UnlockTime 		int64 					`json:"questUnlockTime"`
	ExpireTime 		int64 					`json:"questExpireTime"`
}

type Quest struct {
	DataID 			data.DataId 			`bson:"id"`
	LogicType 		data.QuestLogicType 	`bson:"lt"`
	Progress		map[string]interface{}	`bson:"qp"`
}

func (slot *QuestSlot) MarshalJSON() ([]byte, error) {
	// create client model
	client := &QuestSlotClient{
		UnlockTime: slot.UnlockTime,
		ExpireTime: slot.ExpireTime,
	}

	// client quest 
	quest,_ := json.Marshal(&slot.QuestInstance)
	client.QuestInstance = string(quest)

	// client quest state
	switch(slot.State) {
		case QuestState_Ready:
			client.State = "Ready"
		case QuestState_InProgress:
			client.State = "InProgress"
		case QuestState_Collect:
			client.State = "Collect"
		case QuestState_Cooldown:
			client.State = "Cooldown"
	}

	return json.Marshal(client)
}

func (quest *Quest) MarshalJSON() ([]byte, error) {
	client := map[string]interface{}{}

	client["questDataID"] = data.ToDataName(quest.DataID)

	switch quest.LogicType {

	case data.QuestLogicType_Battle:
		client["logicType"] = "Battle"
		client["numGamesWon"] = quest.Progress["progress"].(int)
		client["chosenCardId"] = quest.Progress["cardId"].(string)

	default:
	}

	return json.Marshal(client)
}

func (slot *QuestSlot) StartCooldown() {
	slot.State = QuestState_Cooldown
	slot.UnlockTime = util.TimeToTicks(time.Now().UTC().Add(data.QuestSlotCooldownTime * time.Minute))
}


func (quest *Quest) UpdateQuest(player *Player) (questComplete bool) {
	switch quest.LogicType {

	case data.QuestLogicType_Battle:
		return quest.UpdateBattleQuest(player)
	default:
	}

	return false
}

func (quest *Quest) UpdateBattleQuest(player *Player) (questComplete bool) {
	questData := data.GetQuestData(quest.DataID)

	//objectives
	completionCondition := questData.Objectives["completionCondition"].(int)
	requiresVictory := questData.Objectives["requiresVictory"].(bool)
	asLeader := questData.Objectives["asLeader"].(bool)

	//progress
	progress := quest.Progress["progress"].(int)
	totalGamesWon := quest.Progress["totalGamesWon"].(int)
	totalGamesPlayed := quest.Progress["totalGamesPlayed"].(int)
	cardId := data.ToDataId(quest.Progress["cardId"].(string)) // will be "" if quest requires no specific card
	noCard := data.ToDataId("")
	currentDeck := player.Decks[player.CurrentDeck]

	// check update conditions and incremement progress if the conditions are met
	if requiresVictory && totalGamesWon < player.WinCount {
		diff := player.WinCount - totalGamesWon

		if cardId == noCard || checkDeckConditions(currentDeck, cardId, asLeader) {
			progress += diff
		}

	} else {
		if totalGamesPlayed < player.MatchCount {
			diff := player.MatchCount - totalGamesPlayed
			
			if cardId == noCard || checkDeckConditions(currentDeck, cardId, asLeader) {
				progress += diff
			}
		}
	}

	// check for quest completion
	questComplete = progress >= completionCondition
	if(questComplete) {
		progress = completionCondition
	}

	quest.Progress["progress"] = progress
	quest.Progress["totalGamesWon"] = player.WinCount
	quest.Progress["totalGamesPlayed"] = player.MatchCount

	return
}

func checkDeckConditions(currentDeck Deck, cardId data.DataId, asLeader bool) bool { // helper func for UpdateBattleQuests
	if asLeader {
		if currentDeck.LeaderCardID == cardId {
			return true
		}
	} else {
		for _,id := range currentDeck.CardIDs {
			if id == cardId {
				return true
			}
		}
	}

	return false
}

func (quest *Quest) IsQuestCompleted() bool {
	switch quest.LogicType {

	case data.QuestLogicType_Battle:
		return quest.IsBattleQuestCompleted()
	default:
	}

	return false
}

func (quest *Quest) IsBattleQuestCompleted() bool {
	questData := data.GetQuestData(quest.DataID)

	completionCondition := questData.Objectives["completionCondition"].(int)
	progress := quest.Progress["progress"].(int)

	return progress >= completionCondition
}

// player specific quest functions below

func (player *Player) SetupQuestDefaults() {
	player.QuestSlots = make([]QuestSlot,3,3)

	player.QuestSlots[0].SupportedTypes = []data.QuestType{data.QuestType_Daily, data.QuestType_Event}
	player.QuestSlots[1].SupportedTypes = []data.QuestType{data.QuestType_Daily, data.QuestType_Event}
	player.QuestSlots[2].SupportedTypes = []data.QuestType{data.QuestType_Weekly}

	for i,_ := range player.QuestSlots {
		player.AssignRandomQuest(i)
	}
}

func (player *Player) CollectQuest(index int, context *util.Context) (*Reward, bool) {
	if player.QuestSlots[index].State != QuestState_Collect && !(&player.QuestSlots[index].QuestInstance).IsQuestCompleted() {
		return nil, false
	}

	//temp
	currency := 100
	questData := data.GetQuestData(player.QuestSlots[index].QuestInstance.DataID)
	if questData.Type == data.QuestType_Weekly {
		currency = 1000
	}

	reward := &Reward{
		StandardCurrency: currency,
	}
	//end temp

	player.QuestSlots[index].StartCooldown()
	player.Save(context)
	return reward, true
}

func (player *Player) AssignRandomQuest(index int, questTypes ...data.QuestType) {
	if len(questTypes) == 0 { // if no quest types have been specified, use the supported types for the slot
		questTypes = player.QuestSlots[index].SupportedTypes
	}

	/* we need to ensure there are no duplicate quests, so build a slice for each populated by complete or in-progress 
		quests and use it in getQuestType to enforce the unique condition */
	currentQuests := make([]data.QuestData,0)
	for i,questSlot := range player.QuestSlots {
		if i != index && (questSlot.State == QuestState_InProgress || questSlot.State == QuestState_InProgress) {
			currentQuests = append(currentQuests, data.GetQuestData(questSlot.QuestInstance.DataID))
		}
	}

	// condition for GetRandomQuestData; we only want unique quests of the type requested for the slot
	getQuest := func(id data.DataId, quest data.QuestData) bool {
		// first iterate through our current quests and ensure we don't pick up a quest with the same objectives
		for _,currentQuest := range currentQuests{ 
			if quest.LogicType == currentQuest.LogicType {
				switch quest.LogicType {
				case data.QuestLogicType_Battle:
					if quest.Objectives["requiresVictory"] == currentQuest.Objectives["requiresVictory"] && 
					   quest.Objectives["asLeader"] == currentQuest.Objectives["asLeader"] {
						return false
					}
				default:
				}
			}
		}

		// last, check to see if this is a supported type of quest
		for _,questType := range questTypes {
			if quest.Type == questType {
				return true
			}
		}
		return false
	}

	questId, questData := data.GetRandomQuestData(getQuest)
	player.AssignQuest(index, questId, questData)
}

func (player *Player) AssignQuest(index int, questId data.DataId, questData data.QuestData) {
	if player.QuestSlots[index].State != QuestState_Ready {
		return 
	}

	player.QuestSlots[index].QuestInstance = Quest {
		DataID: questId,
		LogicType: questData.LogicType,
		Progress: map[string]interface{}{},
	}

	// determine the logic type of the quest and prepare its progress based on the objectives specific to its type
	switch player.QuestSlots[index].QuestInstance.LogicType {

		case data.QuestLogicType_Battle:
			player.QuestSlots[index].QuestInstance.Progress["progress"] = 0
			player.QuestSlots[index].QuestInstance.Progress["totalGamesWon"] = player.WinCount
			player.QuestSlots[index].QuestInstance.Progress["totalGamesPlayed"] = player.MatchCount

			var cardId string
			if questData.Objectives["useRandomCard"].(bool) {
				cardDataId,_ := data.GetRandomCard()
				cardId = data.ToDataName(cardDataId)
			} else {
				cardId = questData.Objectives["cardId"].(string)
			}
			player.QuestSlots[index].QuestInstance.Progress["cardId"] = cardId

		default:
	}

	//assign an expiration date to the slot
	switch questData.Type {

	case data.QuestType_Daily:
		player.QuestSlots[index].ExpireTime = util.TimeToTicks(time.Now().UTC().Add(data.MinutesTillDailyQuestExpires * time.Minute))

	case data.QuestType_Weekly:
		player.QuestSlots[index].ExpireTime = util.TimeToTicks(time.Now().UTC().Add(data.MinutesTillWeeklyQuestExpires * time.Minute))

	default: //events
		// TODO need to identify the event and assign its expiration time to this quest slot
	}

	player.QuestSlots[index].State = QuestState_InProgress
	player.QuestSlots[index].UnlockTime = util.TimeToTicks(time.Now().UTC())
}

// Certain quests types (ex: battle quests) should only update at specific times (ex: immediately after
// a battle), so this function will only update those quests whose logic types are passed as args
func (player *Player) UpdateQuests(context *util.Context, logicTypes ...data.QuestLogicType) error {
	currentTime := util.TimeToTicks(time.Now().UTC())

	//instead of iterating over n logicTypes 3 times (once per slot), add the n logicTypes into a map so
	//we only incur O(1) time per slot to see if its updatable, reducing our total complexity to O(n)
	updatables := map[data.QuestLogicType]int{}
	for i, logicType := range logicTypes {
		updatables[logicType] = i
	}

	for i,_ := range player.QuestSlots {

		if player.QuestSlots[i].State == QuestState_InProgress {
			// check to see if the quest has expired
			if currentTime > player.QuestSlots[i].ExpireTime {
				player.QuestSlots[i].StartCooldown()
				continue
			}

			// check to see if we should update this quest
			if _,updatable := updatables[player.QuestSlots[i].QuestInstance.LogicType]; !updatable {
				continue
			} 

			// call individual update func and check for completion
			if player.QuestSlots[i].QuestInstance.UpdateQuest(player) { // returns true on completion
				player.QuestSlots[i].State = QuestState_Collect
			}
		} else { // check to see if we're ready for a new quest
			if player.QuestSlots[i].State == QuestState_Cooldown && currentTime > player.QuestSlots[i].UnlockTime {
				player.QuestSlots[i].State = QuestState_Ready
				player.AssignRandomQuest(i)
			}
		}
	}

	var err error
	if context != nil {
		err = player.Save(context)
	}
	return err
}