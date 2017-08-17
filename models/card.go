package models

import (
	"encoding/json"

	"bloodtales/data"
)

// server model
type Card struct {
	DataID         			data.DataId   `bson:"id" json:"cardId"`
	Level          			int           `bson:"lv" json:"level"`
	CardCount      			int           `bson:"nm" json:"cardCount"`
	WinCount       			int           `bson:"wc" json:"winCount"`
	LeaderWinCount 			int           `bson:"wl" json:"leaderWinCount"`
}

// client model
type CardClientAlias Card
type CardClient struct {
	DataID                  string        `json:"cardId"`

	*CardClientAlias
}

// custom marshalling
func (card *Card) MarshalJSON() ([]byte, error) {
	// create client model
	client := &CardClient {
		DataID: data.ToDataName(card.DataID),
		CardClientAlias: (*CardClientAlias)(card),
	}
	
	// marshal with client model
	return json.Marshal(client)
}

// custom unmarshalling
func (card *Card) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &CardClient {
		CardClientAlias: (*CardClientAlias)(card),
	}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server data ID
	card.DataID = data.ToDataId(client.DataID)

	return nil
}

func (card *Card) GetDataName() string {
	return data.ToDataName(card.DataID)
}

func (card *Card) GetData() *data.CardData {
	return data.GetCard(card.DataID)
}

func (card *Card) GetPortraitSrc() string {
	data := card.GetData()
	if data != nil {
		return data.GetPortraitSrc()
	}
	return ""
}

func (card *Card) GetPotentialLevel() (level int) {
	level = card.Level
	cardData := card.GetData()
	maxLevel := data.GetMaxLevel(cardData.Rarity)

	progressionData := data.GetCardProgressionData(cardData.Rarity, level)
	cardsNeeded := progressionData.CardsNeeded

	for card.CardCount >= cardsNeeded && level < maxLevel {
		level++
		progressionData = data.GetCardProgressionData(cardData.Rarity, level)
		cardsNeeded += progressionData.CardsNeeded
	}

	return level
}

func (player *Player) AddCards(id data.DataId, num int) {
	//update the card if we already have it, otherwise instantiate a new one and add it in
	for i, card := range player.Cards {
		if card.DataID == id {
			player.Cards[i].CardCount += num
			return
		}
	}

	card := Card {
		DataID: id,
		Level: 1,
		CardCount: num,
		WinCount: 0,
		LeaderWinCount: 0,
	}

	player.Cards = append(player.Cards, card)
}

func (player *Player) GetMapOfCardIndexes() map[data.DataId]int {
	cardMap := map[data.DataId]int {}
	for index, card := range player.Cards {
		cardMap[card.DataID] = index
	}
	return cardMap
}