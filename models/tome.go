package models

import (
	"time"
	"math"
	"encoding/json"
	"math/rand"

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
	DataID         data.DataId   `bson:"id" json:"tomeId"`
	State          TomeState     `bson:"st" json:"state"`
	UnlockTime     time.Time     `bson:"tu" json:"unlockTime"`
}

// client model
type TomeClientAlias Tome
type TomeClient struct {
	DataID         string        `json:"tomeId"`
	State          string        `json:"state"`
	UnlockTime     int64         `json:"unlockTime"`

	*TomeClientAlias
}

// custom marshalling
func (tome *Tome) MarshalJSON() ([]byte, error) {
	// create client model
	client := &TomeClient {
		DataID: data.ToDataName(tome.DataID),
		State: "Locked",
		UnlockTime: util.TimeToTicks(tome.UnlockTime),
		TomeClientAlias: (*TomeClientAlias)(tome),
	}

	// client tome state
	switch tome.State {
	case TomeUnlocking:
		client.State = "Unlocking"
	case TomeUnlocked:
		client.State = "Unlocked"
	}
	
	// marshal with client model
	return json.Marshal(client)
}

// custom unmarshalling
func (tome *Tome) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &TomeClient {
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
	tome.UnlockTime = util.TicksToTime(client.UnlockTime)

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
		return time.Until(tome.UnlockTime).String()
	}
	return "-"
}

func (tome *Tome) GetUnlockCost() int {
	tomeData := data.GetTome(tome.DataID)

	if tome.State != TomeUnlocking {
		return tomeData.GemsToUnlock
	}

	timeNow := util.TimeToTicks(time.Now())
	unlockTime := util.TimeToTicks(tome.UnlockTime) - timeNow
	totalUnlockTime := util.TimeToTicks(time.Now().Add(time.Second * time.Duration(tomeData.TimeToUnlock))) - timeNow

	return int(math.Ceil(float64(tomeData.GemsToUnlock) * float64(unlockTime / totalUnlockTime)))
}

func GetEmptyTome() (tome Tome) {
	tome = Tome{
		DataID: data.ToDataId(""),
		State: TomeEmpty,
		UnlockTime: util.TicksToTime(0),
	}
	return
}

func (tome *Tome) StartUnlocking() {
	tome.State = TomeUnlocking
	tome.UnlockTime = time.Now().Add(time.Duration(data.GetTome(tome.DataID).TimeToUnlock) * time.Second)
}

func (tome *Tome) OpenTome(tier int) (reward *Reward) {
	reward = &Reward{}
	rarities := []string{"COMMON","RARE","EPIC","LEGENDARY"}
	tomeData := data.GetTome(tome.DataID)
	reward.Cards = make([]data.DataId, 0, 6)
	reward.NumRewarded = make([]int, 0, 6)

	for i := 0; i < len(tomeData.GuaranteedRarities); i++ {
		getCards := func(card *data.CardData) bool {
			return card.Rarity == rarities[i] && card.Tier <= tier
		}

		cardSlice := data.GetCards(getCards)

		for j := 0; j < tomeData.GuaranteedRarities[i]; j++ {
			if len(cardSlice) == 0 {
				break
			}

			rand.Seed(time.Now().UTC().UnixNano())
			index := rand.Intn(len(cardSlice))

			card := cardSlice[index]

			if index != (len(cardSlice) - 1) {
				cardSlice[index] = cardSlice[len(cardSlice) - 1]
			} 
			cardSlice = cardSlice[:len(cardSlice) - 1]

			reward.Cards = append(reward.Cards, card)
			reward.NumRewarded = append(reward.NumRewarded, tomeData.CardsRewarded[i])
		}
	}

	rand.Seed(time.Now().UTC().UnixNano())
	reward.PremiumCurrency = tomeData.MinPremiumReward + rand.Intn(tomeData.MaxPremiumReward - tomeData.MinPremiumReward)
	reward.StandardCurrency = tomeData.MinStandardReward + rand.Intn(tomeData.MaxStandardReward - tomeData.MinStandardReward)

	*tome = GetEmptyTome()

	return
}

func (tome *Tome) UpdateTome() {
	if tome.State == TomeUnlocking && time.Now().After(tome.UnlockTime) {
		tome.State = TomeUnlocked
	} 
}

func (player *Player) UpdateTomes(context *util.Context) error {
	unlockTime := util.TicksToTime(player.FreeTomeUnlockTime)

	for time.Now().UTC().After(unlockTime) && player.FreeTomes < 3 {
		unlockTime = unlockTime.Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute)
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

func (player *Player) ModifyArenaPoints(val int) {
	if val < 1 {
		return
	}

	player.ArenaPoints += val

	if player.ArenaPoints > 10 {
		player.ArenaPoints = 10
	}
}

func (player *Player) ClaimTome(context *util.Context, tomeId string) (*Reward, error) {
	tome := &Tome {
		DataID: data.ToDataId(tomeId),
	}

	// check currency
	// TODO

	return player.AddRewards(context, tome)
}

func (player *Player) ClaimFreeTome(context *util.Context,) (tomeReward *Reward, err error) {
	err = player.UpdateTomes(context)

	if player.FreeTomes == 0 || err != nil {
		return
	}

	if player.FreeTomes == 3 {
		player.FreeTomeUnlockTime = util.TimeToTicks(time.Now().Add(time.Duration(MinutesToUnlockFreeTome) * time.Minute))
	}

	player.FreeTomes--

	return player.ClaimTome(context, "TOME_COMMON")
}

func (player *Player) ClaimArenaTome(context *util.Context) (tomeReward *Reward, err error) {
	if player.ArenaPoints < 10 {
		return
	}

	player.ArenaPoints = 0;

	return player.ClaimTome(context, "TOME_RARE")
}