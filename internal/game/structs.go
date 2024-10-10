package game

// Player holds information about a player in the game.
type Player struct {
	ID   string `json:"id"`
	Mark string `json:"mark,omitempty"`
}

// Game represents the state of the game, including the board, current turn, and status.
type Game struct {
	ID     string    `json:"id"`
	Board  [9]string `json:"board"`
	Turn   string    `json:"turn"`
	Winner string    `json:"winner"`
	Status string    `json:"status"`
	Player *Player   `json:"player,omitempty"`
}
