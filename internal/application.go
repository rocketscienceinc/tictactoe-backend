package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rocketscienceinc/tictactoe-backend/internal/config"
	"github.com/rocketscienceinc/tictactoe-backend/internal/repository"
	"github.com/rocketscienceinc/tictactoe-backend/internal/repository/storage"
	"github.com/rocketscienceinc/tictactoe-backend/internal/service"
	"github.com/rocketscienceinc/tictactoe-backend/internal/usecase"
	"github.com/rocketscienceinc/tictactoe-backend/transport/rest"
	"github.com/rocketscienceinc/tictactoe-backend/transport/websocket"
)

var ErrAddrNotFound = errors.New("redis address string is empty")

// RunApp - runs the application.
func RunApp(logger *slog.Logger, conf *config.Config) error {
	log := logger.With("component", "app")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	redisAddrString := conf.Redis.GetRedisAddr()
	if redisAddrString == "" {
		return ErrAddrNotFound
	}

	redisStorage, err := storage.NewRedisStorage(ctx, redisAddrString)
	if err != nil {
		return fmt.Errorf("could not connect to redis storage: %w", err)
	}

	defer func() {
		if err = redisStorage.Connection.Close(); err != nil {
			log.Error("could not close redis storage", "error", err)
		}
	}()

	// create a new sqlite storage
	sqliteStorage, err := storage.NewSQLiteStorage(conf.SQLiteStoragePath)
	if err != nil {
		panic(fmt.Errorf("can't connect to sqlite storage: %w", err))
	}

	// init sqlite storage
	if err = sqliteStorage.Init(ctx); err != nil {
		panic(fmt.Errorf("can't init sqlite storage: %w", err))
	}

	playerRepo := repository.NewPlayerRepository(redisStorage.Connection)
	gameRepo := repository.NewGameRepository(log, redisStorage.Connection)
	userRepo := repository.NewUserRepository(sqliteStorage.Connection)

	playerService := service.NewPlayerService(playerRepo)
	gameService := service.NewGameService(gameRepo)
	botService := service.NewBotService()
	gamePlayService := service.NewGamePlayService(log, playerService, gameService, botService)
	authService := service.NewAuthService(conf.JWTSecretKey)

	gameUseCase := usecase.NewGameUseCase(playerService, gameService, gamePlayService)
	userUseCase := usecase.NewUserUseCase(userRepo)

	pingHandler := rest.NewPingHandler()
	authHandler := rest.NewAuthHandler(log, conf, userUseCase, authService)
	wsHandler := websocket.New(log, gameUseCase)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/ping", pingHandler.PingHandler)
	mux.HandleFunc("/api/auth/google/login", authHandler.GoogleLogin)
	mux.HandleFunc("/api/auth/google/callback", authHandler.GoogleCallback)

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/ws", wsHandler.ServeHTTP)

	srv := &http.Server{
		Addr:         ":" + conf.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Info("Starting HTTP and WebSocket server on", "port", conf.HTTPPort)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	sig := <-sigs
	log.Info("Received signal, shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		panic(fmt.Errorf("server shutdown error: %w", err))
	}

	log.Info("Server gracefully stopped")

	return nil
}
