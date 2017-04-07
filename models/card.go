package models

import (
	"encoding/json"

	"bloodtales/data"
)

type Card struct {
	DataID         data.DataId   `bson:"di" json:"dataId"`
	Level          int           `bson:"lv" json:"level"`
	CardCount      int           `bson:"nm" json:"cardCount"`
	WinCount       int           `bson:"wc" json:"winCount"`
	LeaderWinCount int           `bson:"wl" json:"leaderWinCount"`
}

// custom marshalling
func (card *Card) MarshalJSON() ([]byte, error) {
	type Alias Card
	
	// find card data
	var cardName string = ""
	if cardData, ok := data.GetDataById(card.DataID).(data.CardData); ok {
		cardName = cardData.Name
	} else {
		cardName = "UNKNOWN"
	}
	
	// marshal with client values
	return json.Marshal(&struct {
		DataID string `json:"dataId"`
		*Alias
	}{
		DataID: cardName,
		Alias: (*Alias)(card),
	})
}