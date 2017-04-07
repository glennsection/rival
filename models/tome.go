package models

import (
	"time"

	"bloodtales/data"
)

type TomeState int

const (
	TomeLocked TomeState = iota
	TomeUnlocking
	TomeUnlocked
)

type Tome struct {
	DataID         data.DataId   `bson:"di" json:"dataId"`
	State          TomeState     `bson:"st" json:"state"`
	UnlockTime     time.Time     `bson:"tu" json:"unlockTime"`
}
