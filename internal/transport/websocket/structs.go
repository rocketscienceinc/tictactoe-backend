package websocket

import (
	"encoding/json"

	"github.com/rocketscienceinc/tittactoe-backend/internal/game"
)

// Message represents a WebSocket message with an action type and a payload.
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ResponsePayload struct {
	Player *game.Player  `json:"player"`
	Game   *GameResponse `json:"game,omitempty"`
}

type GameResponse struct {
	ID     string    `json:"id"`
	Board  [9]string `json:"board"`
	Turn   string    `json:"turn"`
	Winner string    `json:"winner"`
	Status string    `json:"status"`
}

// frame represents a WebSocket frame and its metadata.
type frame struct {
	isFin   bool   // Указывает, является ли этот фрейм последним в сообщении
	opCode  byte   // Код операции, указывающий тип данных (например, текстовое сообщение, бинарные данные и т.д.)
	length  uint64 // Длина полезной нагрузки (payload) фрейма
	payload []byte // Данные, передаваемые в фрейме
}
