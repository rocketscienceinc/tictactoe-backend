package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository/storage"
	"github.com/rocketscienceinc/tittactoe-backend/internal/tictactoe"
	"github.com/rocketscienceinc/tittactoe-backend/internal/usecase"
	"github.com/rocketscienceinc/tittactoe-backend/transport/rest"
	"github.com/rocketscienceinc/tittactoe-backend/transport/websocket"
)

var ErrAddrNotFound = errors.New("redis address string is empty")

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info("Received signal, shutting down", "signal", sig)
		cancel()
	}()

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
	gameRepo := repository.NewGameRepository(redisStorage)
	gameController := tictactoe.NewGameController(&entity.Game{})
	gameUseCase := usecase.NewGameManager(logger, playerRepo, gameRepo, gameController)

	// run HTTP server
	httpErrCh := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server", "port", conf.HTTPPort)
		if httpErr := rest.Start(conf.HTTPPort); httpErr != nil {
			log.Error("HTTP server error", "error", httpErr)
			httpErrCh <- httpErr
		}
	}()

	// run Websocket server
	wsErrCh := make(chan error, 1)
	go func() {
		log.Info("Starting WebSocket server", "port", conf.SocketPort)
		wsServer := websocket.New(logger, gameUseCase)
		if wsErr := wsServer.Start(ctx, conf.SocketPort); wsErr != nil {
			log.Error("WebSocket server error", "error", wsErr)
			wsErrCh <- wsErr
		}
	}()

	select {
	case err = <-httpErrCh:
		return fmt.Errorf("HTTP server error: %w", err)
	case err = <-wsErrCh:
		return fmt.Errorf("WebSocket server error: %w", err)
	case <-ctx.Done():
		log.Info("Application context canceled, shutting down")
		return nil
	}
}
