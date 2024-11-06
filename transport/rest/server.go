package rest

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rocketscienceinc/tictactoe-backend/internal/config"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type authService interface {
	GenerateJWTToken(email string) (string, error)
	GenerateStateOauthSession(c echo.Context) (string, error)
}

type userUseCase interface {
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

type Server struct {
	logger *slog.Logger
	conf   *config.Config
}

func NewServer(logger *slog.Logger, conf *config.Config, authService authService, user userUseCase, apiGroup *echo.Group) *Server {
	auth := NewAuth(logger, conf, authService, user)

	apiGroup.GET("/ping", pingHandler)
	apiGroup.GET("/auth/google/login", auth.GoogleLogin)
	apiGroup.GET("/auth/google/callback", auth.GoogleCallback)

	return &Server{
		logger: logger,
		conf:   conf,
	}
}

func pingHandler(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
