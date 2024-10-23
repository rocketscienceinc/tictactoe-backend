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
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository"
	"github.com/rocketscienceinc/tittactoe-backend/internal/repository/storage"
	"github.com/rocketscienceinc/tittactoe-backend/internal/usecase"
	"github.com/rocketscienceinc/tittactoe-backend/transport/rest"
	"github.com/rocketscienceinc/tittactoe-backend/transport/websocket"
)

var ErrAddrNotFound = errors.New("redis address string is empty")

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	ctx, cancel := createAppContext(log)
	defer cancel()

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
	gameUseCase := usecase.NewGameManager(logger, playerRepo, gameRepo)

	httpErrCh := startHTTPServer(log, conf.HTTPPort)
	wsErrCh := startWebSocketServer(ctx, log, conf.SocketPort, gameUseCase)

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

func createAppContext(log *slog.Logger) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info("Received signal, shutting down", "signal", sig)
		cancel()
	}()

	return ctx, cancel
}

func startHTTPServer(log *slog.Logger, port string) chan error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server", "port", port)
		if err := rest.Start(port); err != nil {
			log.Error("HTTP server error", "error", err)
			errCh <- err
		}
	}()

	return errCh
}

func startWebSocketServer(ctx context.Context, log *slog.Logger, port string, gameUseCase *usecase.GameManager) chan error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("Starting WebSocket server", "port", port)
		wsServer := websocket.New(log, gameUseCase)
		if err := wsServer.Start(ctx, port); err != nil {
			log.Error("WebSocket server error", "error", err)
			errCh <- err
		}
	}()

	return errCh
}
