package models

import (
	"time"
	"encoding/json"
	"math/rand"
	"bloodtales/data"
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

//server model
type TomeReward struct {
	Cards 				[]data.DataId
	NumRewarded			[]int 			
	PremiumCurrency 	int 			
	StandardCurrency 	int 			
}

//client model
type TomeRewardClient struct {
	Cards 				[]data.CardData	`json:cards`
	NumRewarded			[]int 			`json:numRewarded` 			
	PremiumCurrency 	int 			`json:PremiumCurrency`		
	StandardCurrency 	int 			`json:StandardCurrency`
}

// custom marshalling
func (tome *Tome) MarshalJSON() ([]byte, error) {
	// create client model
	client := &TomeClient {
		DataID: data.ToDataName(tome.DataID),
		State: "Locked",
		UnlockTime: data.TimeToTicks(tome.UnlockTime),
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

//custom marshalling
func (tomeReward *TomeReward) MarshalJSON() ([]byte, error) {
	//create client model
	client := &TomeRewardClient {
		Cards: make([]data.CardData, len(tomeReward.NumRewarded)),
		NumRewarded: tomeReward.NumRewarded,
		PremiumCurrency: tomeReward.PremiumCurrency,
		StandardCurrency: tomeReward.StandardCurrency,
	}

	for i, id := range tomeReward.Cards {
		client.Cards[i] = *(data.GetCard(id))
	}

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

	if client.DataID == "" {
		tome.State = TomeEmpty
	} else {
		// server tome state
		switch client.State {
		case "Unlocking":
			tome.State = TomeUnlocking
		case "Unlocked":
			tome.State = TomeUnlocked
		default:
			tome.State = TomeLocked
		}
	}

	// server unlock time
	tome.UnlockTime = data.TicksToTime(client.UnlockTime)

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
		data := tome.GetData()
		return (time.Second * time.Duration(data.TimeToUnlock)).String()
	case TomeUnlocking:
		return time.Until(tome.UnlockTime).String()
	}
	return "-"
}

func (tome *Tome) StartUnlocking() {
	tome.State = TomeUnlocking
	tome.UnlockTime = time.Now().Add(time.Duration(data.GetTome(tome.DataID).TimeToUnlock) * time.Second)
}

func (tome *Tome) OpenTome(tier int) (reward *TomeReward) {
	reward = &TomeReward{}
	rarities := []string{"COMMON","RARE","EPIC","LEGENDARY"}
	tomeData := data.GetTome(tome.DataID)
	reward.Cards = make([]data.DataId, 0, 6)
	reward.NumRewarded = make([]int, 0, 6)

	for i := 0; i < len(tomeData.GuaranteedRarities); i++ {
		cardSlice := data.GetCardsByTieredRarity(tier, rarities[i])

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

	tome.DataID = data.ToDataId("")
	tome.State = TomeEmpty
	tome.UnlockTime = data.TicksToTime(0)

	return
}

func (tome *Tome) UpdateTome() {
	if tome.State == TomeUnlocking && time.Now().After(tome.UnlockTime) {
		tome.State = TomeUnlocked
	} 
}