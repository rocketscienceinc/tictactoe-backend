package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/rocketscienceinc/tittactoe-backend/internal/service"
	"github.com/rocketscienceinc/tittactoe-backend/internal/usecase"
	"log/slog"

	"github.com/rocketscienceinc/tittactoe-backend/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"github.com/rocketscienceinc/tittactoe-backend/internal/game"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository/storage"
	"github.com/rocketscienceinc/tittactoe-backend/transport/rest"
	"github.com/rocketscienceinc/tittactoe-backend/transport/websocket"
)

var ErrAddrNotFound = errors.New("redis address string is empty")

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	go runHTTPServer(logger, conf.HTTPPort)

	ctx := context.Background()

	redisAddrString := conf.Redis.GetRedisAddr()
	if redisAddrString == "" {
		return ErrAddrNotFound
	}

	redisStorage, err := storage.New(ctx, redisAddrString)
	if err != nil {
		return fmt.Errorf("could not connect to redis storage: %w", err)
	}

	defer func() {
		if err = redisStorage.Close(); err != nil {
			log.Error("could not close redis storage", "error", err)
		}
	}()

	playerRepo := repository.NewPlayerRepository(redisStorage)

	playerService := service.NewPlayerService(playerRepo)

	gameRepo := repository.NewGameRepository(redisStorage)

	gameService := service.GameService(gameRepo)

	newGame := game.New(entity.Game{})

	gameUsecase := usecase.NewUGame(playerService, gameService, newGame)

	// run Websocket server
	log.Info("Starting WebSocket server", "port:", conf.SocketPort)

	wsServer := websocket.New(logger, gameUsecase)

	if err = wsServer.Start(ctx, conf.SocketPort); err != nil {
		return fmt.Errorf("failed to start socket server: %w", err)
	}

	return nil
}

func runHTTPServer(logger *slog.Logger, port string) {
	log := logger.With("method", "run HTTP server")

	go func() {
		log.Info("Starting HTTP server", "port:", port)
		if err := rest.Start(port); err != nil {
			panic(fmt.Errorf("failed to start HTTP server: %w", err))
		}
	}()
}
