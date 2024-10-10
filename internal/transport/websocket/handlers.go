package websocket

import (
	"bufio"
	"encoding/json"
	"fmt"
)

func (that *Server) handleConnect(msg *Message, bufrw *bufio.ReadWriter) error {
	var payload ResponsePayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal player info: %w", err)
	}

	var responsePayload ResponsePayload

	if payload.Player.ID == "" {
		newPlayerID := GenerateNewSessionID()
		responsePayload = ResponsePayload{
			Player: &Player{
				ID: newPlayerID,
			},
			Game: nil,
		}

		that.logger.Info("registering new player", "player_id", newPlayerID)
	} else {
		responsePayload = ResponsePayload{
			Player: &Player{
				ID: payload.Player.ID,
			},
			Game: nil,
		}
	}

	if err := that.sendMessage(*bufrw, msg.Action, responsePayload); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}
