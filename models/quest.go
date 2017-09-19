package models

import (
	"time"
	"encoding/json"
	
	"bloodtales/data"
	"bloodtales/util"
)

type Quest struct {
	Active 				bool 					`bson:"ac" json:"active"`
	ExpireTime 			int64 					`bson:"ex" json:"expireTime"`
	QuestID				data.DataId 			`bson:"id" json:"-"`
	League				data.League				`bson:"lg" json:"league"`
	Collected			int						`bson:"cl" json:"collected"`
	Properties			map[string]interface{}	`bson:"pp" json:"properties"`
}

type QuestClientAlias Quest
type QuestClient struct {
	QuestID 			string 					`json:"id"`

	*QuestClientAlias
}

// custom marshalling
func (quest *Quest) MarshalJSON() ([]byte, error) {
	// create client model
	client := &QuestClient {
		QuestID:          data.ToDataName(quest.QuestID),

		QuestClientAlias: (*QuestClientAlias)(quest),
	}

	// client uses relative times
	client.ExpireTime = quest.ExpireTime - util.TimeToTicks(time.Now().UTC())

	// marshal with client model
	return json.Marshal(client)
}

func (quest *Quest) UnmarshalJSON(raw []byte) error {
	client := &QuestClient {
		QuestClientAlias: (*QuestClientAlias)(quest),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server data ID
	quest.QuestID = data.ToDataId(client.QuestID)

	return nil
}

func (quest *Quest) GetCurrentObjective() (interface{}) {
	questData := data.GetQuestData(quest.QuestID)
	return questData.Phases[quest.Collected].Objective
}

func (quest *Quest) Update(player *Player) {
	questData := data.GetQuestData(quest.QuestID)

	switch questData.Type {

	case data.QuestTypeSinglePlayerBattle: //must fall through to QuestTypeBattle
		quest.UpdateSinglePlayerBattle(player, questData)
		fallthrough

	case data.QuestTypeBattle:
		quest.UpdateBattle(player, questData)

	case data.QuestTypeTutorial:
		quest.UpdateTutorial(player, questData)

	default:

	}
}

func (quest *Quest) UpdateSinglePlayerBattle(player *Player, questData *data.QuestData) {
	//properties
	requiresVictory := questData.Properties["requiresVictory"].(bool)
	asLeader := questData.Properties["asLeader"].(bool)

	//progress
	progress := quest.Properties["progress"].(int)
	totalPracticeWins := quest.Properties["totalPracticeWins"].(int)
	totalPracticeBattles := quest.Properties["totalPracticeBattles"].(int)
	cardId := data.ToDataId(quest.Properties["cardId"].(string)) // will be "" if quest requires no specific card
	noCard := data.ToDataId("")
	currentDeck := player.Decks[player.CurrentDeck]

	// check update conditions and incremement progress if the conditions are met
	if requiresVictory {
		if totalPracticeWins < player.PracticeWinCount {
			diff := player.PracticeWinCount - totalPracticeWins

			if cardId == noCard || checkDeckConditions(currentDeck, cardId, asLeader) {
				progress += diff
			}
		}

	} else {
		if totalPracticeBattles < player.PracticeMatchCount {
			diff := player.PracticeMatchCount - totalPracticeBattles
			
			if cardId == noCard || checkDeckConditions(currentDeck, cardId, asLeader) {
				progress += diff
			}
		}
	}

	quest.Properties["progress"] = progress
	quest.Properties["totalPracticeWins"] = player.PracticeWinCount
	quest.Properties["totalPracticeBattles"] = player.PracticeMatchCount

	return
}

func (quest *Quest) UpdateBattle(player *Player, questData *data.QuestData) {
	//properties
	requiresVictory := questData.Properties["requiresVictory"].(bool)
	asLeader := questData.Properties["asLeader"].(bool)

	//progress
	progress := quest.Properties["progress"].(int)
	totalGamesWon := quest.Properties["totalGamesWon"].(int)
	totalGamesPlayed := quest.Properties["totalGamesPlayed"].(int)
	cardId := data.ToDataId(quest.Properties["cardId"].(string)) // will be "" if quest requires no specific card
	noCard := data.ToDataId("")
	currentDeck := player.Decks[player.CurrentDeck]

	// check update conditions and incremement progress if the conditions are met
	if requiresVictory {
		if totalGamesWon < player.WinCount {
			diff := player.WinCount - totalGamesWon

			if cardId == noCard || checkDeckConditions(currentDeck, cardId, asLeader) {
				progress += diff
			}
		}

	} else {
		if totalGamesPlayed < player.MatchCount {
			diff := player.MatchCount - totalGamesPlayed
			
			if cardId == noCard || checkDeckConditions(currentDeck, cardId, asLeader) {
				progress += diff
			}
		}
	}

	quest.Properties["progress"] = progress
	quest.Properties["totalGamesWon"] = player.WinCount
	quest.Properties["totalGamesPlayed"] = player.MatchCount

	return
}

func (quest *Quest) UpdateTutorial(player *Player, questData *data.QuestData) {
	progress := quest.Properties["progress"].(int)
	currentObjective := quest.GetCurrentObjective().(string)

	for i := progress; progress < len(questData.Phases); progress++ {
		if player.TutorialCompleted(currentObjective) {
			progress = i
		} else {
			break
		}
	}

	quest.Properties["progress"] = progress
}

func checkDeckConditions(currentDeck Deck, cardId data.DataId, asLeader bool) bool { // helper func for UpdateBattleQuests
	if asLeader {
		if currentDeck.LeaderCardID == cardId {
			return true
		}
	} else {
		for _, id := range currentDeck.CardIDs {
			if id == cardId {
				return true
			}
		}
	}

	return false
}

func (quest *Quest) IsCollectable() bool {
	questData := data.GetQuestData(quest.QuestID)

	steps := len(questData.Phases)
	if quest.Collected < steps {
		switch(questData.Type) {
		case data.QuestTypeSinglePlayerBattle: //must fall through to QuestTypeBattle
			fallthrough

		case data.QuestTypeBattle:
			currentObjective := quest.GetCurrentObjective().(int)
			progress := quest.Properties["progress"].(int)
			return progress >= currentObjective

		case data.QuestTypeTutorial:
			progress := quest.Properties["progress"].(int)
			return progress > quest.Collected

		}

	}
	return false
}

func (quest *Quest) IsCompleted() bool {
	questData := data.GetQuestData(quest.QuestID)

	steps := len(questData.Phases)
	return quest.Collected >= steps
}

// player specific quest functions below

func (player *Player) SetupQuestDefaults() {
	type StartingQuest struct {
		QuestID 		data.DataId
		Index 			int
	}

	startingQuests := make([]StartingQuest, 0)

	if len(player.Quests) != 0 {
		for i := range player.Quests {
			if player.Quests[i].Active {
				startingQuests = append(startingQuests, StartingQuest {
					QuestID: player.Quests[i].QuestID,
					Index: i,
					})
			}
		}
	}

	player.Quests = make([]Quest, 3, 3)

	for i := range startingQuests {
		questData := data.GetQuestData(startingQuests[i].QuestID)
		player.AssignQuest(startingQuests[i].Index, startingQuests[i].QuestID, questData)
	}

	for i := range player.Quests {
		if !player.Quests[i].Active {
			player.AssignRandomQuest(i)
		}
	}
}

func (player *Player) CollectQuest(index int, context *util.Context) (*Reward, bool, error) {
	quest := &player.Quests[index]

	if !quest.Active || !quest.IsCollectable() {
		return nil, false, nil
	}

	questData := data.GetQuestData(quest.QuestID)
	reward := player.GetReward(questData.Phases[quest.Collected].RewardID, quest.League, player.GetLevel())

	quest.Collected += 1

	// check if this is a progressive quest
	if quest.IsCompleted() {
		quest.Active = false

		//permanent daily quests should expire at midnight the day after they're collected, so set their expiration date now
		if questData.Permanent  {
			if questData.Period == data.QuestPeriodDailyPermanent {
				quest.ExpireTime = util.TimeToTicks(util.GetDateInNDays(player.TimeZone, 1))
			} else { //permanent weekly quests should expire at midnight on the next weekly reset
				if questData.Period == data.QuestPeriodWeeklyPermanent {
					expirationDate := util.GetDateOfNextWeekday(player.TimeZone, time.Monday, false)
					quest.ExpireTime = util.TimeToTicks(expirationDate)
				}
			}
		} 
	}

	err := player.AddRewards(reward, context)

	return reward, true, err
}

func (player *Player) AssignRandomQuest(index int) {
	if player.GetLevel() < data.GetQuestSlotMinLevel(index) { //Player isn't high enough level for a quest in this slot yet
		return
	}

	/* we need to ensure there are no duplicate quests, so build a slice of quests regardless of their current state 
		and use it in getQuestPeriod to enforce the unique condition */
	currentQuestDatas := make([]*data.QuestData, 0)
	for _, quest := range player.Quests {
		if quest.Active {
			currentQuestDatas = append(currentQuestDatas, data.GetQuestData(quest.QuestID))
		}
	}

	// condition for GetRandomQuestData; we only want unique quests of the type requested for the slot
	getQuest := func(id data.DataId, questData *data.QuestData) bool {
		// check percent chance (TODO - actually factor in percent chance if > 0)
		if questData.PercentChance <= 0 {
			return false
		}

		// check support
		if !questData.IsSupported(index) {
			return false
		}

		// first iterate through our current quests and ensure we don't pick up a quest with the same objectives
		for _, currentQuestData := range currentQuestDatas { 
			if currentQuestData != nil && questData.Type == currentQuestData.Type {
				switch questData.Type {

				case data.QuestTypeSinglePlayerBattle: //must fall through to QuestTypeBattle
					fallthrough

				case data.QuestTypeBattle:
					if questData.Period == currentQuestData.Period {
						if questData.Properties["requiresVictory"] == currentQuestData.Properties["requiresVictory"] && 
						   questData.Properties["asLeader"] == currentQuestData.Properties["asLeader"] {
							return false
						}
					}

				default:

				}
			}
		}

		return true
	}

	questId, questData := data.GetRandomQuestData(getQuest)
	player.AssignQuest(index, questId, questData)
}

func (player *Player) AssignQuest(index int, questId data.DataId, questData *data.QuestData) {
	quest := &player.Quests[index]

	if quest.Active {
		// cannot override active quest
		return 
	}

	// init basic info
	quest.Active = true
	quest.QuestID = questId
	quest.League = data.GetLeague(data.GetRank(player.RankPoints).Level)
	quest.Collected = 0
	quest.Properties = map[string]interface{}{}

	// determine the logic type of the quest and prepare its progress based on the objectives specific to its type
	switch questData.Type {
	case data.QuestTypeSinglePlayerBattle: //must fall through to QuestTypeBattle
		quest.Properties["totalPracticeWins"] = player.PracticeWinCount
		quest.Properties["totalPracticeBattles"] = player.PracticeMatchCount
		fallthrough

	case data.QuestTypeBattle:
		quest.Properties["progress"] = 0
		quest.Properties["totalGamesWon"] = player.WinCount
		quest.Properties["totalGamesPlayed"] = player.MatchCount

		var cardId string
		if questData.Properties["useRandomCard"].(bool) {
			cardId = data.ToDataName(player.Cards[util.RandomIntn(len(player.Cards))].DataID)
		} else {
			cardId = questData.Properties["cardId"].(string)
		}
		quest.Properties["cardId"] = cardId
		break

	case data.QuestTypeTutorial:
		quest.Properties["progress"] = 0
		break

	default:
		
	}

	//assign an expiration date to the slot
	if !questData.Permanent { //permanent quests do not expire until quest has been completed
		switch questData.Period {

		case data.QuestPeriodDaily:
			quest.ExpireTime = util.TimeToTicks(util.GetDateInNDays(player.TimeZone, 1)) //get tomorrow's date

		case data.QuestPeriodWeekly:
			expirationDate := util.GetDateOfNextWeekday(player.TimeZone, time.Monday, false)
			quest.ExpireTime = util.TimeToTicks(expirationDate)

		default: //events
			// TODO need to identify the event and assign its expiration time to this quest slot
		}
	}
}

// Certain quests types (ex: battle quests) should only update at specific times (ex: immediately after
// a battle), so this function will only update those quests whose logic types are passed as args
func (player *Player) UpdateQuests(context *util.Context, questTypes ...data.QuestType) error {
	currentTime := util.TimeToTicks(time.Now().UTC())

	//instead of iterating over n questTypes 3 times (once per slot), add the n questTypes into a map so
	//we only incur O(1) time per slot to see if its updatable
	updatables := map[data.QuestType]int{}
	for i, questType := range questTypes {
		updatables[questType] = i
	}

	for i, _ := range player.Quests {
		quest := &player.Quests[i]
		questData := data.GetQuestData(quest.QuestID)

		if quest.Active {
			// check to see if the quest has expired. if so, assign a new quest
			if !questData.Permanent && currentTime > quest.ExpireTime {
				quest.Active = false
				player.AssignRandomQuest(i)
				continue
			}

			// check to see if we should update this quest
			if _, updatable := updatables[questData.Type]; !updatable {
				continue
			} 

			// call individual update func and check for completion
			quest.Update(player)
		} else { // check to see if we're ready for a new quest
			if currentTime > quest.ExpireTime {
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