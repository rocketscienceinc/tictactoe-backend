package socket

import "encoding/json"

// Message represents a WebSocket message with an action type and a payload.
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ResponsePayload struct {
	Player *Player `json:"player,omitempty"`
	Game   *Game   `json:"game,omitempty"`
}

type JoinGamePayload struct {
	Player struct {
		ID string `json:"id"`
	} `json:"player"`
	Room struct {
		ID string `json:"id"`
	} `json:"room"`
}

type TurnPayload struct {
	Player struct {
		ID string `json:"id"`
	} `json:"player"`
	Game struct {
		ID string `json:"id"`
	} `json:"game"`
	Cell int `json:"cell"`
}

// Player holds information about a player in the game.
type Player struct {
	ID   string `json:"id"`
	Mark string `json:"mark,omitempty"`
}

// Game represents the state of the game, including the board, current turn, and status.
type Game struct {
	ID      string    `json:"id"`
	Board   [9]string `json:"board"`
	Turn    string    `json:"turn"`
	Winner  string    `json:"winner"`
	Status  string    `json:"status"`
	Players []Player  `json:"players,omitempty"`
}

// frame represents a WebSocket frame and its metadata.
type frame struct {
	isFin   bool   // Указывает, является ли этот фрейм последним в сообщении
	opCode  byte   // Код операции, указывающий тип данных (например, текстовое сообщение, бинарные данные и т.д.)
	length  uint64 // Длина полезной нагрузки (payload) фрейма
	payload []byte // Данные, передаваемые в фрейме
}
