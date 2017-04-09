package models

import (
	"time"
	"encoding/json"

	"bloodtales/data"
)

// tome state
type TomeState int
const (
	TomeLocked TomeState = iota
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

	// server tome state
	switch client.State {
	case "Unlocking":
		tome.State = TomeUnlocking
	case "Unlocked":
		tome.State = TomeUnlocked
	default:
		tome.State = TomeLocked
	}

	// server unlock time
	tome.UnlockTime = data.TicksToTime(client.UnlockTime)

	return nil
}