package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type gameUseCase interface {
	GetOrCreatePlayer(ctx context.Context, playerID string) (*entity.Player, error)
	GetOrCreateGame(ctx context.Context, playerID, gameType string) (*entity.Game, error)
	GetGameByPlayerID(ctx context.Context, playerID string) (*entity.Game, error)
	JoinGameByID(ctx context.Context, gameID, playerID string) (*entity.Game, error)
	JoinWaitingPublicGame(ctx context.Context, playerID string) (*entity.Game, error)
	MakeTurn(ctx context.Context, playerID string, cell int) (*entity.Game, error)
}

type Server struct {
	logger      *slog.Logger
	gameUseCase gameUseCase

	handlers    map[string]func(ctx context.Context, message *IncomingMessage, ws *websocket.Conn) error
	connections map[string]*websocket.Conn
	rooms       map[string][2]Player
	connMutex   sync.RWMutex
}

type Player struct {
	conn *websocket.Conn
	id   string
}

func New(logger *slog.Logger, gameUseCase gameUseCase) *Server {
	server := &Server{
		logger:      logger,
		gameUseCase: gameUseCase,
		handlers:    make(map[string]func(context.Context, *IncomingMessage, *websocket.Conn) error),
		connections: make(map[string]*websocket.Conn),
		rooms:       make(map[string][2]Player),
	}

	server.handlers["connect"] = server.handleConnect
	server.handlers["game:new"] = server.handleNewGame
	server.handlers["game:join"] = server.handleJoinGame
	server.handlers["game:turn"] = server.handleGameTurn

	server.rooms["1"] = [2]Player{
		{id: "1"},
		{id: "2"},
	}

	server.rooms["2"] = [2]Player{
		{id: "5"},
		{id: "4"},
	}

	return server
}

func (that *Server) HandleConnection(c echo.Context) error {
	log := that.logger.With("method", "HandleConnection")
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		log.Info("Websocket connected")

		currentPlayerID := ""

		for roomID, players := range that.rooms {
			for _, player := range players {
				if player.id == currentPlayerID {
					fmt.Println("player room id:", roomID)
					break
				}
			}
		}

		that.connections[ws.Request().RemoteAddr] = ws
		for _, conn := range that.connections {
			if conn.Request().RemoteAddr == ws.Request().RemoteAddr {
				continue
			}

			if err := websocket.Message.Send(conn, fmt.Sprintf("client %s was connected", ws.Request().RemoteAddr)); err != nil {
				log.Error("Failed to send welcome message", "error", err)
			}
		}

		that.handleMessages(c.Request().Context(), ws)
	})
	wsHandler.ServeHTTP(c.Response(), c.Request())

	return nil
}

// handleMessages - processes messages from the client.
// ToDo: Need to move user creation to rest api, in order to track connections and remove user from connections.
func (that *Server) handleMessages(ctx context.Context, ws *websocket.Conn) {
	log := that.logger.With("method", "HandleMessages")

	for {
		var message IncomingMessage
		if err := websocket.JSON.Receive(ws, &message); err != nil {
			// if client closed connection
			if errors.Is(err, io.EOF) {
				log.Info("websocket connection closed by client")
				break
			}

			// if network error
			if errors.Is(err, net.ErrClosed) {
				log.Info("network connection closed")
				break
			}

			log.Error("failed to receive message", "error", err)
			break
		}

		handler, ok := that.handlers[message.Action]
		if !ok {
			log.Error("action handler not found", "action", message.Action)
			if err := that.sendError(ws, message.Action, fmt.Sprintf("action handler not found action: %s", message.Action)); err != nil {
				break
			}
			continue
		}

		if err := handler(ctx, &message, ws); err != nil {
			log.Error("invalid handle message", "error", err)
			continue
		}
	}
}

func (that *Server) sendMessage(ws *websocket.Conn, action string, payload Payload) error {
	response := OutgoingMessage{
		Action:  action,
		Payload: payload,
	}

	if err := websocket.JSON.Send(ws, response); err != nil {
		that.logger.Error("failed to send message", "error", err)
		return err
	}
	return nil
}

func (that *Server) sendError(ws *websocket.Conn, action, errorMsg string) error {
	payload := Payload{
		Error: errorMsg,
	}
	return that.sendMessage(ws, action, payload)
}

type OutgoingMessage struct {
	Action  string  `json:"action"`
	Payload Payload `json:"payload,omitempty"`
}

type IncomingMessage struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Payload struct {
	Player *entity.Player `json:"player,omitempty"`
	Game   *entity.Game   `json:"game,omitempty"`
	Error  string         `json:"error,omitempty"`
	Cell   int            `json:"cell,omitempty"`
}
