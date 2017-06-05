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

type Quest struct {
	DataID 			data.DataId 			`bson:"id"`
	LogicType 		data.QuestLogicType 	`bson:"lt"`
	Progress		util.Stream 			`bson:"qp"`
}

func (quest *Quest) MarshalJSON() ([]byte, error) {
	client := map[string]interface{}{}

	client["questDataID"] = data.ToDataName(quest.DataID)

	switch quest.LogicType {

	case data.QuestLogicType_Battle:
		client["logicType"] = "Battle"
		client["numGamesWon"] = quest.Progress.GetRequiredInt("progress")
		client["chosenCardId"] = data.ToDataName(data.DataId(quest.Progress.GetRequiredInt64("cardId")))

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
	completionCondition := questData.Objectives.GetRequiredInt("completionCondition")
	requiresVictory := questData.Objectives.GetRequiredBool("requiresVictory")
	winAsLeader := questData.Objectives.GetRequiredBool("winAsLeader")

	//progress
	progress := quest.Progress.GetRequiredInt("progress")
	totalGamesWon := quest.Progress.GetRequiredInt("totalGamesWon")
	totalGamesPlayed := quest.Progress.GetRequiredInt("totalGamesPlayed")
	cardId := data.DataId(quest.Progress.GetRequiredInt64("cardId"))

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

	quest.Progress.Set("progress", progress)
	quest.Progress.Set("totalGamesWon", player.WinCount)
	quest.Progress.Set("totalGamesPlayed", player.MatchCount)

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
	quest := Quest {
		DataID: questId,
		LogicType: questData.LogicType,
		Progress: *data.NewQuestObjectivesStreamSource(),
	}

	// determine the logic type of the quest and prepare its progress based on the objectives specific to its type
	switch quest.LogicType {

		case data.QuestLogicType_Battle:
			quest.Progress.Set("progress", 0)
			quest.Progress.Set("totalGamesWon", player.WinCount)
			quest.Progress.Set("totalGamesPlayed", player.MatchCount)
			
			if questData.Objectives.GetRequiredBool("useRandomCard") {
				//Get a random card
			} else {
				quest.Progress.Set("cardId", data.DataId(questData.Objectives.GetRequiredInt64("cardId")))
			}

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