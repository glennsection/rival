package models

import (
	"encoding/json"
	"math"
	"time"

	"bloodtales/data"
	"bloodtales/util"
)

// tome state
type TomeState int

const (
	TomeEmpty TomeState = iota
	TomeLocked
	TomeUnlocking
	TomeUnlocked
)

// server model
type Tome struct {
	DataID     data.DataId `bson:"id" json:"tomeId"`
	State      TomeState   `bson:"st" json:"state"`
	UnlockTime int64       `bson:"tu" json:"unlockTime"`
	League 	   data.League `bson:"cl" json:"-"`
}

// client model
type TomeClientAlias Tome
type TomeClient struct {
	DataID     string `json:"tomeId"`
	State      string `json:"state"`
	UnlockTime int64  `json:"unlockTime"`
	League 	   string `json:"league"`

	*TomeClientAlias
}

// custom marshalling
func (tome *Tome) MarshalJSON() ([]byte, error) {
	// create client model
	client := &TomeClient{
		DataID:          data.ToDataName(tome.DataID),
		State:           "Locked",
		UnlockTime:      tome.UnlockTime - util.TimeToTicks(time.Now().UTC()),
		TomeClientAlias: (*TomeClientAlias)(tome),
	}

	// client tome state
	switch tome.State {
	case TomeUnlocking:
		client.State = "Unlocking"
	case TomeUnlocked:
		client.State = "Unlocked"
	}

	if client.DataID != "INVALID" {
		client.League = data.GetLeagueData(tome.League).ID
	}

	// marshal with client model
	return json.Marshal(client)
}

// custom unmarshalling
func (tome *Tome) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &TomeClient{
		TomeClientAlias: (*TomeClientAlias)(tome),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server data ID
	tome.DataID = data.ToDataId(client.DataID)

	   // server tome state
    switch client.State {
    case "Unlocking":
        tome.State = TomeUnlocking
    case "Unlocked":
        tome.State = TomeUnlocked
    case "Locked":
        tome.State = TomeLocked
    default:
        tome.State = TomeEmpty
    }

    // server unlock time
    tome.UnlockTime = client.UnlockTime

    // server league
    tome.League = data.GetLeagueByID(client.League)

    return nil
}

func (tome *Tome) GetDataName() string {
	return data.ToDataName(tome.DataID)
}

func (tome *Tome) GetData() *data.TomeData {
	return data.GetTome(tome.DataID)
}

func (tome *Tome) GetImageSrc() string {
	data := tome.GetData()
	if data != nil {
		return data.GetImageSrc()
	}
	return "/static/img/tomes/tome_NONE.png"
}

func (tome *Tome) GetStateName() string {
	switch tome.State {
	default:
		return "Empty"
	case TomeLocked:
		return "Locked"
	case TomeUnlocking:
		return "Unlocking"
	case TomeUnlocked:
		return "Unlocked"
	}
}

func (tome *Tome) GetUnlockRemaining() string {
	switch tome.State {
	case TomeLocked:
		return (time.Second * time.Duration(tome.GetData().TimeToUnlock)).String()
	case TomeUnlocking:
		return time.Until(util.TicksToTime(tome.UnlockTime)).String()
	}
	return "-"
}

func (tome *Tome) GetUnlockCost() int {
	tomeData := data.GetTome(tome.DataID)

	costMultiplier := data.GetLeagueData(tome.League).TomeCostMultiplier

	if tome.State != TomeUnlocking {
		return int(float64(tomeData.GemsToUnlock) * costMultiplier)
	}

	timeRemaining := util.TicksToTime(tome.UnlockTime).Sub(time.Now().UTC())
	return int(math.Ceil(float64(tomeData.GemsToUnlock) * costMultiplier * (timeRemaining.Seconds() / float64(tomeData.TimeToUnlock))))
}

func GetEmptyTome() (tome Tome) {
	tome = Tome{
		DataID:     data.ToDataId(""),
		State:      TomeEmpty,
		UnlockTime: 0,
		League: 	data.LeagueOne,
	}
	return
}

func (player *Player) GetEmptyTomeSlot() (index int, tome *Tome) {
	tome = nil
	index = -1

	for i, tomeSlot := range player.Tomes {
		if tomeSlot.State == TomeEmpty {
			tome = &player.Tomes[i]
			index = i
			break
		}
	}

	return
}

func (tome *Tome) StartUnlocking() {
	tome.State = TomeUnlocking
	tome.UnlockTime = util.TimeToTicks(time.Now().Add(time.Duration(data.GetTome(tome.DataID).TimeToUnlock) * time.Second))
}

func (player *Player) OpenTome(tome *Tome) (reward *Reward) {
	tomeData := data.GetTome(tome.DataID)

	reward = player.GetReward(tomeData.RewardID, tome.League, player.GetLevel())

	*tome = GetEmptyTome()

	return
}

func (tome *Tome) UpdateTome() {
	if tome.State == TomeUnlocking && time.Now().After(util.TicksToTime(tome.UnlockTime)) {
		tome.State = TomeUnlocked
	}
}

func (player *Player) UpdateTomes(context *util.Context) error {
	unlockTime := util.TicksToTime(player.FreeTomeUnlockTime)

	for time.Now().UTC().After(unlockTime) && player.FreeTomes < 3 {
		unlockTime = unlockTime.Add(time.Duration(data.GameplayConfig.FreeTomeUnlockTime) * time.Second)
		player.FreeTomes++
	}

	player.FreeTomeUnlockTime = util.TimeToTicks(unlockTime)

	for i, _ := range player.Tomes {
		(&player.Tomes[i]).UpdateTome()
	}

	var err error
	if context != nil {
		err = player.Save(context)
	}

	return err
}

func (player *Player) AddTomeRewards(context *util.Context, tome *Tome) (reward *Reward, err error) {
	reward = player.OpenTome(tome)
	err = player.AddRewards(reward, context)
	return
}

func (player *Player) ModifyArenaPoints(val int) {
	if val < 1 || time.Now().UTC().Before(util.TicksToTime(player.ArenaTomeUnlockTime)) {
		return
	}

	player.ArenaPoints += val

	if player.ArenaPoints > 10 {
		player.ArenaPoints = 10
	}
}

func (player *Player) ClaimFreeTome(context *util.Context) (tomeReward *Reward, err error) {
	err = player.UpdateTomes(context)

	if player.FreeTomes == 0 || err != nil {
		return
	}

	if player.FreeTomes == 3 {
		player.FreeTomeUnlockTime = util.TimeToTicks(time.Now().Add(time.Duration(data.GameplayConfig.FreeTomeUnlockTime) * time.Second))
	}

	player.FreeTomes--

	tomeReward = player.GetReward(data.ToDataId("TOME_FREE_REWARD"), data.GetLeague(data.GetRank(player.RankPoints).Level), player.GetLevel())
	err = player.AddRewards(tomeReward, context)

	return
}

func (player *Player) ClaimArenaTome(context *util.Context) (tomeReward *Reward, err error) {
	if player.ArenaPoints < 10 || time.Now().UTC().Before(util.TicksToTime(player.ArenaTomeUnlockTime)) {
		return
	}

	player.ArenaPoints = 0
	player.ArenaTomeUnlockTime = util.TimeToTicks(time.Now().UTC().Add(time.Duration(data.GameplayConfig.BattleTomeCooldown) * time.Second))

	tomeReward = player.GetReward(data.ToDataId("TOME_BATTLE_REWARD"), data.GetLeague(data.GetRank(player.RankPoints).Level), player.GetLevel())
	err = player.AddRewards(tomeReward, context)

	return
}
