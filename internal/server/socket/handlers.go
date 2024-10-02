package socket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// handleSessionCookie - handles user session.
func (that *Server) handleSessionCookie(writer http.ResponseWriter, req *http.Request, log *slog.Logger) {
	cookie, err := req.Cookie("user_session")
	if err != nil {
		cookie = &http.Cookie{
			Name:    "user_session",
			Value:   GenerateNewSessionID(),
			Expires: time.Now().Add(24 * time.Hour),
			Path:    "/ws",
		}
		http.SetCookie(writer, cookie)
		log.Info("session cookie not found, new one created", "cookie", cookie.Value)
	} else {
		log.Info("session cookie found", "cookie", cookie.Value)
	}
}

// HandleMessages - processes messages from the client.
func (that *Server) HandleMessages(bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "HandleMessages")

	for {
		msg, err := that.readMessage(bufrw)
		if err != nil {
			log.Error("error reading message", "error", err)
			return err
		}

		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Error("failed to unmarshal message", "error", err)
			continue
		}

		if err := that.processMessage(&message, bufrw); err != nil {
			log.Error("error processing message", "error", err)
		}
	}
}

func (that *Server) handleConnect(msg *Message, bufrw *bufio.ReadWriter) error {
	var payload PlayerPayload

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal player info: %w", err)
	}

	var responsePayload ResponsePayload

	if payload.Player.ID == "" {
		newPlayerID := GenerateNewSessionID()
		responsePayload = ResponsePayload{
			Player: PlayerInfo{
				ID: newPlayerID,
			},
			Game: nil,
		}

		that.logger.Info("registering new player", "player_id", newPlayerID)
	} else {
		responsePayload = ResponsePayload{
			Player: PlayerInfo{
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
