package models

import (
	"encoding/json"

	"bloodtales/data"
)

type Card struct {
	DataID         data.DataId   `bson:"id" json:"cardId"`
	Level          int           `bson:"lv" json:"level"`
	CardCount      int           `bson:"nm" json:"cardCount"`
	WinCount       int           `bson:"wc" json:"winCount"`
	LeaderWinCount int           `bson:"wl" json:"leaderWinCount"`
}

// custom marshalling
func (card *Card) MarshalJSON() ([]byte, error) {
	type Alias Card
	
	// convert to client data names
	var cardName string = data.ToDataName(card.DataID)
	if cardName == "" {
		cardName = "UNKNOWN"
	}
	
	// marshal with client values
	return json.Marshal(&struct {
		DataID string `json:"cardId"`
		*Alias
	}{
		DataID: cardName,
		Alias: (*Alias)(card),
	})
}

// custom unmarshalling
func (card *Card) UnmarshalJSON(raw []byte) error {
	type Alias Card

	// temp struct
	aux := &struct {
		DataID string `json:"cardId"`
		*Alias
	}{
		Alias: (*Alias)(card),
	}

	if err := json.Unmarshal(raw, &aux); err != nil {
		return err
	}

	// convert to server values
	card.DataID = data.ToDataId(aux.DataID)

	return nil
}