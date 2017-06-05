package models

import (
	"time"
	"encoding/json"
	
	"gopkg.in/mgo.v2"

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
	Quest 			Quest 					`bson:"qi"`
	State 			QuestState 				`bson:"qs"`
	UnlockTime 		int64 					`bson:"ut"`
	ExpireTime 		int64 					`bson:"et"`
}

type QuestSlotClient struct {
	Quest 			string 					`json:"questInstance"`
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
	quest,_ := json.Marshal(&slot.Quest)
	client.Quest = string(quest)

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
	slot.UnlockTime = util.TimeToTicks(time.Now().Add(data.QuestSlotCooldownTime * time.Minute))
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
	winAsLeader := questData.Objectives["winAsLeader"].(bool)

	//progress
	progress := quest.Progress["progress"].(int)
	totalGamesWon := quest.Progress["totalGamesWon"].(int)
	totalGamesPlayed := quest.Progress["totalGamesPlayed"].(int)
	cardId := data.ToDataId(quest.Progress["cardId"].(string))

	// check update conditions and incremement progress if the conditions are met
	if requiresVictory && totalGamesWon < player.WinCount {
		diff := player.WinCount - totalGamesWon

		if !winAsLeader {
			for _,id := range player.Decks[player.CurrentDeck].CardIDs {
				if id == cardId {
					progress += diff
				}
			}
		} else {
			if player.Decks[player.CurrentDeck].LeaderCardID == cardId {
				progress += diff
			}
		}
	} else {
		if totalGamesPlayed < player.MatchCount {
			diff := player.MatchCount - totalGamesPlayed
			progress += diff
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

// player specific quest functions below

func (player *Player) CollectQuest(index int, database *mgo.Database) (*Reward, bool) {
	if player.Quests[index].State != QuestState_Collect {
		return nil, false
	}

	//temp
	reward := &Reward{
		StandardCurrency: 100,
	}

	player.Quests[index].StartCooldown()
	player.Save(database)
	return reward, true
}

func (player *Player) AssignRandomQuest(slot *QuestSlot) {
	if slot.State != QuestState_Ready {
		return 
	}

	// prepare our BaseQuestData with an identifier
	questId, questData := data.GetRandomQuestData()
	slot.Quest = Quest {
		DataID: questId,
		LogicType: questData.LogicType,
		Progress: map[string]interface{}{},
	}

	// determine the logic type of the quest and prepare its progress based on the objectives specific to its type
	switch slot.Quest.LogicType {

		case data.QuestLogicType_Battle:
			slot.Quest.Progress["progress"] = 0
			slot.Quest.Progress["totalGamesWon"] = player.WinCount
			slot.Quest.Progress["totalGamesPlayed"] = player.MatchCount

			var cardId string
			if questData.Objectives["useRandomCard"].(bool) {
				cardDataId,_ := data.GetRandomCard()
				cardId = data.ToDataName(cardDataId)
			} else {
				cardId = questData.Objectives["cardId"].(string)
			}
			slot.Quest.Progress["cardId"] = cardId

		default:
	}

	//assign an expiration date to the slot
	switch questData.Type {

	case data.QuestType_Daily:
		slot.ExpireTime = util.TimeToTicks(time.Now().Add(data.MinutesTillDailyQuestExpires * time.Minute))

	case data.QuestType_Weekly:
		slot.ExpireTime = util.TimeToTicks(time.Now().Add(data.MinutesTillWeeklyQuestExpires * time.Minute))

	default: //events
		// TODO need to identify the event and assign its expiration time to this quest slot
	}

	slot.State = QuestState_InProgress
	slot.UnlockTime = util.TimeToTicks(time.Now().UTC())
}

// Certain quests types (ex: battle quests) should only update at specific times (ex: immediately after
// a battle), so this function will only update those quests whose logic types are passed as args
func (player *Player) UpdateQuests(logicTypes ...data.QuestLogicType) {
	currentTime := util.TimeToTicks(time.Now().UTC())

	//instead of iterating over n logicTypes 3 times (once per slot), add the n logicTypes into a map so
	//we only incur O(1) time per slot to see if its updatable, reducing our total complexity to O(n)
	updatables := map[data.QuestLogicType]int{}
	for i, logicType := range logicTypes {
		updatables[logicType] = i
	}

	for _, slot := range player.Quests {

		if slot.State == QuestState_InProgress {
			// check to see if the quest has expired
			if currentTime > slot.ExpireTime {
				slot.StartCooldown()
				continue
			}

			// check to see if we should update this quest
			if _,updatable := updatables[slot.Quest.LogicType]; !updatable {
				continue
			}

			// call individual update func and check for completion
			if slot.Quest.UpdateQuest(player) { // returns true on completion
				slot.State = QuestState_Collect
			}
		} else { // check to see if we're ready for a new quest
			if slot.State == QuestState_Cooldown && currentTime > slot.UnlockTime {
				slot.State = QuestState_Ready
				player.AssignRandomQuest(&slot)
			}
		}
	}
}