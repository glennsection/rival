package models

type Deck struct {
	LeaderCardID int   `bson:"ld" json:"leaderCardId"`
	CardIDs      []int `bson:"cd" json:"cardIds"`
}

func (deck *Deck) SetDeckCard(card int, deckIndex int) {
	//card is already at its desired position
	if card == deck.CardIDs[deckIndex] {
		return
	}

	//card is the current leader card, swap
	if card == deck.LeaderCardID {
		deck.LeaderCardID = deck.CardIDs[deckIndex]
		deck.CardIDs[deckIndex] = card
		return
	}

	//card is already in the deck, swap
	for _, deckCard := range deck.CardIDs {
		if deckCard == card {
			deckCard = deck.CardIDs[deckIndex]
			deck.CardIDs[deckIndex] = card
			return
		}
	}

	//the card is not in the deck
	deck.CardIDs[deckIndex] = card
}

func (deck *Deck) SetLeaderCard(card int) {
	//card is already the leader card
	if card == deck.LeaderCardID {
		return
	}

	//card is in the deck, swap
	for _, deckCard := range deck.CardIDs {
		if deckCard == card {
			deckCard = deck.LeaderCardID
			deck.LeaderCardID = card
			return
		}
	}

	//the card is not in the deck
	deck.LeaderCardID = card
}
