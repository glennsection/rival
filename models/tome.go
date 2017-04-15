package models

import (
	"time"
	"encoding/json"

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

func (tome *Tome) OpenTome() {
	tome.DataID = data.ToDataId("")
	tome.State = TomeEmpty
	tome.UnlockTime = data.TicksToTime(0)
}

func (tome *Tome) UpdateTome() {
	if tome.State == TomeUnlocking && time.Now().After(tome.UnlockTime) {
		tome.State = TomeUnlocked
	} 
}