package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/net/websocket"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

func (that *Server) handleConnect(ctx context.Context, msg *IncomingMessage, ws *websocket.Conn) error {
	log := that.logger.With("method", "handleConnect")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	player, err := that.gameUseCase.GetOrCreatePlayer(ctx, payloadReq.Player.ID)
	if err != nil {
		log.Error("failed to create or get", "player", err)
		return that.sendError(ws, msg.Action, "failed to create a new player")
	}

	that.connMutex.Lock()
	that.connections[player.ID] = ws
	that.connMutex.Unlock()

	if player.GameID != "" {
		game, err := that.gameUseCase.GetGameByPlayerID(ctx, player.ID)
		if err != nil {
			log.Info("failed to get the game", "game", player.GameID)
			return that.sendError(ws, msg.Action, "failed to get the game")
		}

		payloadResp := Payload{
			Player: player,
			Game:   game,
		}
		payloadResp.Game.Players = nil
		payloadResp.Game.Type = ""

		if err = that.sendMessage(ws, msg.Action, payloadResp); err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}

		return nil
	}

	payloadResp := Payload{
		Player: player,
	}

	if err = that.sendMessage(ws, msg.Action, payloadResp); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("successfully connected player")

	return nil
}

func (that *Server) handleNewGame(ctx context.Context, msg *IncomingMessage, ws *websocket.Conn) error {
	log := that.logger.With("method", "handleNewGame")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	that.connMutex.Lock()
	that.connections[payloadReq.Player.ID] = ws
	that.connMutex.Unlock()

	game, err := that.gameUseCase.GetOrCreateGame(ctx, payloadReq.Player.ID, payloadReq.Game.Type)
	if err != nil {
		log.Error("failed to create or get", "player", err)
		return that.sendError(ws, msg.Action, "failed to create a new game")
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		that.connMutex.Lock()
		conn, ok := that.connections[player.ID]
		that.connMutex.Unlock()

		if !ok {
			log.Error("failed to find connection")
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   game,
		}

		payloadResp.Game.Players = nil
		payloadResp.Game.Type = ""

		if err = that.sendMessage(conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player is already in game")

	return nil
}

func (that *Server) handleJoinGame(ctx context.Context, msg *IncomingMessage, ws *websocket.Conn) error {
	log := that.logger.With("method", "handleJoinGame")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	that.connMutex.Lock()
	that.connections[payloadReq.Player.ID] = ws
	that.connMutex.Unlock()

	log = log.With("playerID", payloadReq.Player.ID)

	var game *entity.Game
	var err error

	if payloadReq.Game.IsPublic() {
		game, err = that.gameUseCase.JoinWaitingPublicGame(ctx, payloadReq.Player.ID) // ToDo: Nee to move usecase
		if err != nil {
			log.Info("failed to join public game", "error", err)
			return that.sendError(ws, msg.Action, fmt.Sprintf("game %s: %v", payloadReq.Game.ID, err))
		}
	}

	if !payloadReq.Game.IsPublic() {
		game, err = that.gameUseCase.JoinGameByID(ctx, payloadReq.Game.ID, payloadReq.Player.ID)
		if err != nil {
			log.Error("failed to join game", "error", err)
			return that.sendError(ws, msg.Action, fmt.Sprintf("game %s: %v", payloadReq.Game.ID, err))
		}
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		that.connMutex.Lock()
		conn, ok := that.connections[player.ID]
		that.connMutex.Unlock()

		if !ok {
			log.Info("failed to find connection")
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   game,
		}

		payloadResp.Game.Players = nil
		payloadResp.Game.Type = ""

		if err = that.sendMessage(conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player joined game")

	return nil
}

func (that *Server) handleGameTurn(ctx context.Context, msg *IncomingMessage, ws *websocket.Conn) error {
	log := that.logger.With("method", "handleGameTurn")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	that.connMutex.Lock()
	that.connections[payloadReq.Player.ID] = ws
	that.connMutex.Unlock()

	log = log.With("playerID", payloadReq.Player.ID)

	game, err := that.gameUseCase.MakeTurn(ctx, payloadReq.Player.ID, payloadReq.Cell)
	if errors.Is(err, apperror.ErrGameFinished) {
		if err = that.handleGameFinished(msg.Action, game); err != nil {
			return that.sendError(ws, msg.Action, fmt.Sprintf("failed to finish game %s: %v", game.ID, err))
		}

		return nil
	}

	log = log.With("gameID", game.ID)

	if errors.Is(err, apperror.ErrGameIsNotStarted) {
		return that.sendError(ws, msg.Action, fmt.Sprintf("game %s: %v", game.ID, err))
	}

	if errors.Is(err, apperror.ErrCellOccupied) {
		return that.sendError(ws, msg.Action, fmt.Sprintf("game %s: %v", game.ID, err))
	}

	if err != nil {
		log.Error("failed to make turn", "error", err)
		return that.sendError(ws, msg.Action, fmt.Sprintf("failed to turn in game %v", err))
	}

	for _, player := range game.Players {
		that.connMutex.Lock()
		conn, ok := that.connections[player.ID]
		that.connMutex.Unlock()

		if !ok {
			log.Error("failed to find connection")
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   game,
		}

		payloadResp.Game.Players = nil
		payloadResp.Game.Type = ""

		if err = that.sendMessage(conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player made a turn")

	return nil
}

func (that *Server) handleGameFinished(action string, game *entity.Game) error {
	log := that.logger.With("method", "handleGameFinished")

	payloadResp := Payload{
		Game: game,
	}

	for _, player := range game.Players {
		if player.IsBot() {
			continue
		}

		that.connMutex.Lock()
		conn, ok := that.connections[player.ID]
		that.connMutex.Unlock()

		if !ok {
			log.Error("failed to find connection", "player", player.ID)
			continue
		}

		payloadResp.Game.Players = nil
		payloadResp.Game.Type = ""

		if err := that.sendMessage(conn, action, payloadResp); err != nil {
			return fmt.Errorf("failed to send game finished message %s: %w", player.ID, err)
		}
	}

	log.Info("Game finished", "gameID", game.ID)

	return nil
}
