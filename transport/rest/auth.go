package rest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/rocketscienceinc/tictactoe-backend/internal/config"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

const urlUserInfo = "https://www.googleapis.com/oauth2/v2/userinfo"

type AuthHandler interface {
	GoogleLogin(ctx echo.Context) error
	GoogleCallback(ctx echo.Context) error
}

type authHandler struct {
	logger *slog.Logger

	oauthConfig      *oauth2.Config
	oauthStateString string

	auth authService
	user userUseCase
}

func NewAuth(logger *slog.Logger, conf *config.Config, auth authService, user userUseCase) AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     conf.GoogleOAuth.ClientID,
		ClientSecret: conf.GoogleOAuth.ClientSecret,

		RedirectURL: conf.GoogleOAuth.RedirectURL,

		Scopes:   conf.GoogleOAuth.Scopes,
		Endpoint: google.Endpoint,
	}

	return &authHandler{
		logger:      logger.With(),
		oauthConfig: oauthConfig,
		auth:        auth,
		user:        user,
	}
}

func (that *authHandler) GoogleLogin(ctx echo.Context) error {
	stateToken, err := that.auth.GenerateStateOauthSession(ctx)
	if err != nil {
		that.logger.Error("Failed to generate state token", "error", err)
		return ctx.String(http.StatusInternalServerError, "Internal Server Error")
	}

	// generate authURL for authorization with session token.
	authURL := that.oauthConfig.AuthCodeURL(stateToken)
	return ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (that *authHandler) GoogleCallback(ctx echo.Context) error {
	log := that.logger.With("method", "GoogleCallBack")

	// get state from session.
	userSession, err := session.Get("session", ctx)
	if err != nil {
		log.Error("failed to get session", "error", err)
		return ctx.String(http.StatusInternalServerError, "Internal Server Error")
	}

	// check existence of the state and it`s type.
	storedState, ok := userSession.Values["state"].(string)
	if !ok || storedState == "" {
		log.Error("state not found in session")
		return ctx.String(http.StatusBadRequest, "Invalid session state")
	}

	// get state and code from the request parameters.
	state := ctx.QueryParam("state")
	code := ctx.QueryParam("code")

	// check if state matches.
	if state != storedState {
		log.Error("invalid OAuth state", "expected", storedState, "got", state)
		return ctx.String(http.StatusBadRequest, "Invalid OAuth state")
	}

	// exchange code for token.
	token, err := that.oauthConfig.Exchange(ctx.Request().Context(), code)
	if err != nil {
		log.Error("failed to exchange code for token", "error", err)
		return ctx.String(http.StatusInternalServerError, "Internal Server Error")
	}

	// getting user information
	client := that.oauthConfig.Client(ctx.Request().Context(), token)
	userInfo, err := getUserInfo(client)
	if err != nil {
		log.Error("failed to get user info", "error", err)
		return ctx.String(http.StatusInternalServerError, "Internal Server Error")
	}

	user, err := that.user.Update(ctx.Request().Context(), userInfo)
	if err != nil {
		log.Error("failed to create or update user", "error", err)
		return ctx.String(http.StatusInternalServerError, "Internal Server Error")
	}

	jwtToken, err := that.auth.GenerateJWTToken(user.Email)
	if err != nil {
		log.Error("failed to generate JWT token", "error", err)
		return ctx.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"token": jwtToken,
	})
}

func getUserInfo(client *http.Client) (*entity.User, error) {
	resp, err := client.Get(urlUserInfo)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo entity.User
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
