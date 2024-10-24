package entity

type Player struct {
	ID     string `json:"id,omitempty"`
	Mark   string `json:"mark,omitempty"`
	GameID string `json:"game_id,omitempty"`
}
