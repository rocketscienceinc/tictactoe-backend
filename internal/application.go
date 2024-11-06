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

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/rocketscienceinc/tictactoe-backend/internal/config"
	"github.com/rocketscienceinc/tictactoe-backend/internal/repository"
	"github.com/rocketscienceinc/tictactoe-backend/internal/repository/storage"
	"github.com/rocketscienceinc/tictactoe-backend/internal/repository/storage/sqlite"
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

	redisStorage, err := storage.New(ctx, redisAddrString)
	if err != nil {
		return fmt.Errorf("could not connect to redis storage: %w", err)
	}

	defer func() {
		if err = redisStorage.Connection.Close(); err != nil {
			log.Error("could not close redis storage", "error", err)
		}
	}()

	// create a new sqlite storage
	sqliteStorage, err := sqlite.New(conf.SQLiteStoragePath)
	if err != nil {
		panic(fmt.Errorf("can't connect to sqlite storage: %w", err))
	}

	// init sqlite storage
	if err = sqliteStorage.Init(ctx); err != nil {
		panic(fmt.Errorf("can't init sqlite storage: %w", err))
	}

	userRepo := repository.NewUserRepository(sqliteStorage.Connection)
	playerRepo := repository.NewPlayerRepository(redisStorage.Connection)
	gameRepo := repository.NewGameRepository(log, redisStorage.Connection)

	playerService := service.NewPlayerService(playerRepo)
	gameService := service.NewGameService(gameRepo)
	botService := service.NewBotService()
	gamePlayService := service.NewGamePlayService(log, playerService, gameService, botService)

	gameUseCase := usecase.NewGameUseCase(playerService, gameService, gamePlayService)

	authService := service.NewAuthService(conf.Security.SecretKey)
	userUseCase := usecase.NewUserUseCase(userRepo)

	e := echo.New()

	e.Use(middleware.Logger())

	// secretKey := []byte(conf.Security.SecretKey)
	//if len(secretKey) < 32 {
	//	return fmt.Errorf("secret key must be at least 32 bytes long")
	//}
	//
	//store := sessions.NewCookieStore(secretKey)
	//store.Options = &sessions.Options{
	//	Path:     "/",
	//	MaxAge:   86400 * 1,
	//	HttpOnly: true,
	//	Secure:   false,
	//	SameSite: http.SameSiteStrictMode,
	//}

	apiGroup := e.Group("/api")
	// apiGroup.Use(session.Middleware(sessions.NewCookieStore([]byte(conf.Security.SecretKey))))

	rest.NewServer(log, conf, authService, userUseCase, apiGroup) // ToDo: Register()
	wsHandlers := websocket.New(log, gameUseCase)

	e.GET("/ws", wsHandlers.HandleConnection) // ToDo: NOT WORKED

	go func() {
		addr := fmt.Sprintf(":%s", conf.HTTPPort)
		log.Info("Starting HTTP and WebSocket server on", "port", conf.HTTPPort)
		if err = e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	sig := <-sigs
	log.Info("Received signal, shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err = e.Shutdown(shutdownCtx); err != nil {
		panic(fmt.Errorf("server shutdown error: %w", err))
	}

	log.Info("Server gracefully stopped")

	return nil
}
