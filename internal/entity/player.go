package entity

import "strings"

type Player struct {
	ID     string `json:"id,omitempty"`
	Mark   string `json:"mark,omitempty"`
	GameID string `json:"game_id,omitempty"`
}

func NewBotPlayer(gameID string, mark string) *Player {
	return &Player{
		ID:     "bot:" + gameID,
		Mark:   mark,
		GameID: gameID,
	}
}

func (that *Player) IsBot() bool {
	return strings.HasPrefix(that.ID, "bot:")
}
