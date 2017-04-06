package models

import (
	"bloodtales/data"
)

type Deck struct {
	LeaderCardID data.DataId   `bson:"ld" json:"leaderCardId"`
	CardIDs      []data.DataId `bson:"cd" json:"cardIds"`
}
