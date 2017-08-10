package models

import (
	"encoding/json"

	"bloodtales/data"
)

type Deck struct {
	LeaderCardID 	data.DataId 	`bson:"ld" json:"leaderCardId"`
	CardIDs      	[]data.DataId  	`bson:"cd" json:"cardIds"`
}

type DeckClientAlias struct {
	LeaderCardID 	string 			`json:"leaderCardId"`
	CardIDs 		[]string 		`json:"cardIds"`
}

// custom marshalling
func (deck *Deck) MarshalJSON() ([]byte, error) {
	// create client model
	client := &DeckClientAlias {
		LeaderCardID: data.ToDataName(deck.LeaderCardID),
	}

	client.CardIDs = make([]string, len(deck.CardIDs))
	for i, cardId := range deck.CardIDs {
		client.CardIDs[i] = data.ToDataName(cardId)
	}
	
	// marshal with client model
	return json.Marshal(client)
}

// custom unmarshalling
func (deck *Deck) UnmarshalJSON(raw []byte) error {
	// create client model
	client := &DeckClientAlias {}

	// unmarshal to client model
	if err := json.Unmarshal(raw, &client); err != nil {
		return err
	}

	// server data IDs
	deck.LeaderCardID = data.ToDataId(client.LeaderCardID)

	deck.CardIDs = make([]data.DataId, len(client.CardIDs))
	for i, cardId := range client.CardIDs {
		deck.CardIDs[i] = data.ToDataId(cardId)
	}

	return nil
}

func (deck *Deck) SetDeckCard(card data.DataId, deckIndex int) {
	//card is already at its desired position
	if card == deck.CardIDs[deckIndex] {
		return
	}

	//card is the current leader card, swap
	if card == deck.LeaderCardID {
		deck.LeaderCardID = deck.CardIDs[deckIndex]
	} else {
		//card is already in the deck, swap
		for i, deckCard := range deck.CardIDs {
			if deckCard == card {
				deck.CardIDs[i] = deck.CardIDs[deckIndex]
				break
			}
		}
	}

	//the card is not in the deck
	deck.CardIDs[deckIndex] = card
}

func (deck *Deck) SetLeaderCard(card data.DataId) {
	//card is already the leader card
	if card == deck.LeaderCardID {
		return
	}

	//card is in the deck, swap
	for i, deckCard := range deck.CardIDs {
		if deckCard == card {
			deck.CardIDs[i] = deck.LeaderCardID
			break
		}
	}

	//the card is not in the deck
	deck.LeaderCardID = card
}
