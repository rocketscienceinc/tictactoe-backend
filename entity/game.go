package entity

type Game struct {
	ID         string    `json:"id"`
	Board      [9]string `json:"board"`
	Winner     string    `json:"winner"`
	Status     string    `json:"status"`
	PlayerTurn string    `json:"player_turn"`
	Players    []*Player `json:"players"`
}
