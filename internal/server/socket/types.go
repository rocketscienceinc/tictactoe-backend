package socket

import "encoding/json"

// Message represents a WebSocket message with an action type and a payload.
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// PlayerInfo holds information about a player in the game.
type PlayerInfo struct {
	ID   string `json:"id"`
	Mark string `json:"mark,omitempty"`
}

// Game represents the state of the game, including the board, current turn, and status.
type Game struct {
	ID     string   `json:"id"`
	Board  []string `json:"board"`
	Turn   string   `json:"turn"`
	Winner string   `json:"winner"`
	Status string   `json:"status"`
}

type ResponsePayload struct {
	Player PlayerInfo `json:"player"`
	Game   *Game      `json:"game,omitempty"`
}

type PlayerPayload struct {
	Player struct {
		ID string `json:"id"`
	} `json:"player"`
}

// frame represents a WebSocket frame and its metadata.
type frame struct {
	isFin   bool   // Указывает, является ли этот фрейм последним в сообщении
	opCode  byte   // Код операции, указывающий тип данных (например, текстовое сообщение, бинарные данные и т.д.)
	length  uint64 // Длина полезной нагрузки (payload) фрейма
	payload []byte // Данные, передаваемые в фрейме
}
