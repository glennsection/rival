package models

import (
	"bloodtales/data"
)

type Card struct {
	//ID             bson.ObjectId `bson:"_id,omitempty" json:"id"`
	DataID         data.DataId   `bson:"di" json:"dataId"`
	Level          int           `bson:"lv" json:"level"`
	CardCount      int           `bson:"nm" json:"cardCount"`
	WinCount       int           `bson:"wc" json:"winCount"`
	LeaderWinCount int           `bson:"wl" json:"leaderWinCount"`
}
