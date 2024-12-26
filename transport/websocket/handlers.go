package websocket

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

const (
	gameStatusOpponentOut  = "opponent_out"
	payloadActionGameLeave = "game:leave"
	gameStatusLeave        = "leave"
)

func (that *Server) handleConnect(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleConnect")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payloadReq.Player == nil {
		log.Error("Player is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Player is required")
	}

	player, err := that.gameUseCase.GetOrCreatePlayer(ctx, payloadReq.Player.ID)
	if err != nil {
		log.Error("failed to create or get", "player", err)

		return that.sendErrorResponse(bufrw, msg.Action, "failed to create a new player")
	}

	that.connectionsMutex.Lock()
	that.connections[player.ID] = bufrw
	that.connectionsMutex.Unlock()

	that.playerReconnected(player.ID)

	if player.GameID != "" {
		return that.handleExistingGame(ctx, bufrw, msg, player)
	}

	payloadResp := Payload{
		Player: player,
	}

	if err = that.sendMessage(bufrw, msg.Action, payloadResp); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Info("successfully connected player")

	return nil
}

// handleExistingGame processes a player already in a game.
func (that *Server) handleExistingGame(ctx context.Context, bufrw *bufio.ReadWriter, msg *Message, player *entity.Player) error {
	log := that.logger.With("method", "handleExistingGame")

	game, err := that.gameUseCase.GetGameByPlayerID(ctx, player.ID)
	if err != nil {
		log.Error("failed to get game", "gameID", player.GameID, "error", err)
		return that.sendErrorResponse(bufrw, msg.Action, "failed to get the game")
	}

	payload := Payload{
		Player: player,
		Game:   maskGameDetails(game),
	}

	return that.sendMessage(bufrw, msg.Action, payload)
}

func (that *Server) handleNewGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleNewGame")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payloadReq.Player == nil {
		log.Error("Player is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Player is required")
	}

	if payloadReq.Game == nil {
		log.Error("Game is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Game is required")
	}

	that.connectionsMutex.Lock()
	that.connections[payloadReq.Player.ID] = bufrw
	that.connectionsMutex.Unlock()

	var game *entity.Game
	var err error

	if payloadReq.Game.IsPublic() {
		game, err = that.gameUseCase.CreateOrJoinToPublicGame(ctx, payloadReq.Player.ID, payloadReq.Game.Type)
		if err != nil {
			log.Error("failed to create or join to public game", "game", payloadReq.Game.Type)
			return that.sendErrorResponse(bufrw, msg.Action, "failed to create or join to public game")
		}
	}

	if !payloadReq.Game.IsPublic() {
		game, err = that.gameUseCase.GetOrCreateGame(ctx, payloadReq.Player.ID, payloadReq.Game.Type)
		if err != nil {
			log.Error("failed to create or get", "player", err)
			return that.sendErrorResponse(bufrw, msg.Action, "failed to create a new game")
		}
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		if player.IsBot() {
			continue
		}

		that.connectionsMutex.RLock()
		conn, ok := that.connections[player.ID]
		that.connectionsMutex.RUnlock()

		if !ok {
			log.Warn("connection not found for player", "playerID", player.ID)
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   maskGameDetails(game),
		}

		if err = that.sendMessage(conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player is already in game")

	return nil
}

func (that *Server) handleJoinGame(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleJoinGame")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	if payloadReq.Player == nil {
		log.Error("Player is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Player is required")
	}

	if payloadReq.Game == nil {
		log.Error("Game is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Game is required")
	}

	that.connectionsMutex.Lock()
	that.connections[payloadReq.Player.ID] = bufrw
	that.connectionsMutex.Unlock()

	log = log.With("playerID", payloadReq.Player.ID)

	game, err := that.gameUseCase.JoinGameByID(ctx, payloadReq.Game.ID, payloadReq.Player.ID)
	if err != nil {
		log.Error("failed to join game", "error", err)
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s: %v", payloadReq.Game.ID, err))
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		if player.IsBot() {
			continue
		}

		that.connectionsMutex.RLock()
		conn, ok := that.connections[player.ID]
		that.connectionsMutex.RUnlock()

		if !ok {
			log.Info("failed to find connection")
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   maskGameDetails(game),
		}

		if err = that.sendMessage(conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player joined game")

	return nil
}

func (that *Server) handleGameTurn(ctx context.Context, msg *Message, bufrw *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleGameTurn")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	if payloadReq.Player == nil {
		log.Error("Player is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Player is required")
	}

	if payloadReq.Cell == nil {
		log.Error("Game is missing in payload")
		return that.sendErrorResponse(bufrw, msg.Action, "Game is required")
	}

	that.connectionsMutex.Lock()
	that.connections[payloadReq.Player.ID] = bufrw
	that.connectionsMutex.Unlock()

	log = log.With("playerID", payloadReq.Player.ID)

	game, err := that.gameUseCase.MakeTurn(ctx, payloadReq.Player.ID, *payloadReq.Cell)
	if errors.Is(err, apperror.ErrGameFinished) {
		if err = that.handleGameFinished(msg.Action, game); err != nil {
			return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("failed to finish game %s: %v", game.ID, err))
		}

		return nil
	}

	if errors.Is(err, apperror.ErrGameIsNotStarted) {
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s: %v", game.ID, err))
	}

	if errors.Is(err, apperror.ErrCellOccupied) {
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("game %s: %v", game.ID, err))
	}

	if err != nil {
		log.Error("failed to make turn", "error", err)
		return that.sendErrorResponse(bufrw, msg.Action, fmt.Sprintf("failed to turn in game %v", err))
	}

	log = log.With("gameID", game.ID)

	for _, player := range game.Players {
		that.connectionsMutex.RLock()
		conn, ok := that.connections[player.ID]
		that.connectionsMutex.RUnlock()

		if !ok {
			log.Error("failed to find connection")
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   maskGameDetails(game),
		}

		if err = that.sendMessage(conn, msg.Action, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player made a turn")

	return nil
}

func (that *Server) handleGameLeave(ctx context.Context, msg *Message, bufRW *bufio.ReadWriter) error {
	log := that.logger.With("method", "handleGameLeave")

	var payloadReq Payload

	if err := json.Unmarshal(msg.Payload, &payloadReq); err != nil {
		return fmt.Errorf("failed to unmarshal playload: %w", err)
	}

	if payloadReq.Player == nil {
		log.Error("Player is missing in payload")
		return that.sendErrorResponse(bufRW, msg.Action, "Player is required")
	}

	that.connectionsMutex.Lock()
	that.connections[payloadReq.Player.ID] = bufRW
	that.connectionsMutex.Unlock()

	game, err := that.gameUseCase.GetGameByPlayerID(ctx, payloadReq.Player.ID)
	if err != nil {
		log.Error("failed to find game", "error", err)
		return that.sendErrorResponse(bufRW, msg.Action, "game doesn't exist")
	}

	err = that.gameUseCase.EndGame(ctx, game)
	if err != nil {
		log.Error("failed to end game", "error", err)
		return that.sendErrorResponse(bufRW, msg.Action, "game doesn't exist")
	}

	for _, player := range game.Players {
		if player.IsBot() {
			continue
		}

		that.connectionsMutex.Lock()
		conn, ok := that.connections[player.ID]
		that.connectionsMutex.Unlock()

		if !ok {
			log.Info("failed to find connection")
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   maskGameDetails(game),
		}

		payloadResp.Game.Status = gameStatusLeave

		if err = that.sendMessage(conn, payloadActionGameLeave, payloadResp); err != nil {
			log.Error("failed to send game update", "error", err)
		}
	}

	log.Info("Player leaving")

	return nil
}

func (that *Server) handleGameFinished(action string, game *entity.Game) error {
	log := that.logger.With("method", "handleGameFinished")

	for _, player := range game.Players {
		if player.IsBot() {
			continue
		}

		that.connectionsMutex.RLock()
		conn, ok := that.connections[player.ID]
		that.connectionsMutex.RUnlock()

		if !ok {
			log.Error("failed to find connection", "player", player.ID)
			continue
		}

		payloadResp := Payload{
			Player: player,
			Game:   maskGameDetails(game),
		}

		if err := that.sendMessage(conn, action, payloadResp); err != nil {
			return fmt.Errorf("failed to send game finished message %s: %w", player.ID, err)
		}
	}

	log.Info("Game finished", "gameID", game.ID)

	return nil
}

func (that *Server) handleDisconnect(bufRW *bufio.ReadWriter) {
	log := that.logger.With("method", "handleDisconnect")

	that.connectionsMutex.Lock()
	var disconnectedPlayerID string
	for playerID, connection := range that.connections {
		if connection == bufRW {
			disconnectedPlayerID = playerID
			break
		}
	}

	if disconnectedPlayerID == "" {
		log.Error("disconnected player not found in connections")
		that.connectionsMutex.Unlock()
		return
	}

	delete(that.connections, disconnectedPlayerID)
	log.Info("player disconnected", "playerID", disconnectedPlayerID)
	that.connectionsMutex.Unlock()

	that.disconnectedMutex.Lock()
	that.disconnectedPlayers[disconnectedPlayerID] = time.Now()
	that.disconnectedMutex.Unlock()
}

func (that *Server) handleOpponentOut(ctx context.Context, playerID string) {
	log := that.logger.With("method", "handleOpponentOut")

	game, err := that.gameUseCase.GetGameByPlayerID(ctx, playerID)
	if err != nil {
		log.Error("failed to get game by player ID", "playerID", playerID, "error", err)
		return
	}

	if game == nil {
		log.Error("game not found for player", "playerID", playerID)
		return
	}

	err = that.gameUseCase.EndGame(ctx, game)
	if err != nil {
		log.Error("failed to finish game", "gameID", game.ID, "error", err)
		return
	}

	for _, player := range game.Players {
		if player.ID == playerID || player.IsBot() {
			continue
		}

		that.connectionsMutex.RLock()
		opponentConn, ok := that.connections[player.ID]
		that.connectionsMutex.RUnlock()

		if !ok {
			log.Warn("opponent connection not found", "playerID", player.ID)
			continue
		}

		payloadResp := Payload{
			Game: maskGameDetails(game),
		}
		payloadResp.Game.Status = gameStatusOpponentOut

		if err = that.sendMessage(opponentConn, payloadActionGameLeave, payloadResp); err != nil {
			log.Error("failed to send game:leave message", "playerID", player.ID, "error", err)
		}
	}

	log.Info("handled opponent out", "gameID", game.ID)
}

func (that *Server) playerReconnected(playerID string) {
	that.disconnectedMutex.Lock()
	defer that.disconnectedMutex.Unlock()
	delete(that.disconnectedPlayers, playerID)
}

// maskGameDetails hides sensitive details from the game payload.
func maskGameDetails(game *entity.Game) *entity.Game {
	game.Players = nil
	game.Type = ""
	return game
}

func (that *Server) sendErrorResponse(bufrw *bufio.ReadWriter, action, errorMsg string) error {
	payload := Payload{Error: errorMsg}
	if err := that.sendMessage(bufrw, action, payload); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	return nil
}
