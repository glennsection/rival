package models

// create new player data
func (player *Player) Initialize() {
	player.StandardCurrency = 1000
	player.PremiumCurrency = 10
	player.XP = 0
	player.Cards = nil
	player.Decks = nil
	player.CurrentDeck = 0
}
