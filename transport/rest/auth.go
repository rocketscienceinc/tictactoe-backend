package rest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/rocketscienceinc/tictactoe-backend/internal/config"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

const (
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

	cookieSecure         = true
	cookieHTTPOnly       = true
	cookieNameOAuthState = "oauthstate"
	cookieNameJWTToken   = "auth_token"
	cookieExpireDuration = 24 * time.Hour
	oauthCookieExpire    = 365 * 24 * time.Hour

	formKeyCode  = "code"
	formKeyState = "state"

	rootURLPath = "/"
)

type AuthHandler interface {
	GoogleLogin(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
}

type userUseCase interface {
	GetOrCreate(ctx context.Context, user *entity.User) (*entity.User, error)
}

type authService interface {
	GenerateJWT(userID string) (string, error)
}

type authHandler struct {
	logger *slog.Logger

	oauthConfig *oauth2.Config

	user userUseCase
	auth authService
}

func NewAuthHandler(logger *slog.Logger, conf *config.Config, user userUseCase, auth authService) AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     conf.GoogleOAuth.ClientID,
		ClientSecret: conf.GoogleOAuth.ClientSecret,

		RedirectURL: conf.GoogleOAuth.RedirectURL,

		Scopes:   conf.GoogleOAuth.Scopes,
		Endpoint: google.Endpoint,
	}
	return &authHandler{
		logger:      logger.With("component", "auth"),
		oauthConfig: oauthConfig,
		user:        user,
		auth:        auth,
	}
}

func (that *authHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	log := that.logger.With("method", "GoogleLogin")
	stateToken, err := that.generateStateOauthCookie(w)
	if err != nil {
		log.Error("Error generating state token", "error", err)
		http.Redirect(w, r, rootURLPath, http.StatusTemporaryRedirect)
		return
	}

	authURL := that.oauthConfig.AuthCodeURL(stateToken)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (that *authHandler) GoogleCallback(writer http.ResponseWriter, req *http.Request) {
	log := that.logger.With("method", "GoogleCallBack")

	// read oauthState from Cookie.
	oauthState, _ := req.Cookie(cookieNameOAuthState)

	if req.FormValue(formKeyState) != oauthState.Value {
		log.Error("invalid oauth google state")
		http.Redirect(writer, req, rootURLPath, http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := that.getUserDataFromGoogle(req.Context(), req.FormValue(formKeyCode))
	if err != nil {
		log.Error("failed to get user data", "error", err)
		http.Redirect(writer, req, rootURLPath, http.StatusTemporaryRedirect)
		return
	}

	user, err := that.user.GetOrCreate(req.Context(), userInfo)
	if err != nil {
		log.Error("failed to get or create user", "error", err)
		http.Redirect(writer, req, rootURLPath, http.StatusTemporaryRedirect)
		return
	}

	jwtToken, err := that.auth.GenerateJWT(user.ID)
	if err != nil {
		log.Error("failed to generate JWT", "error", err)
		http.Redirect(writer, req, rootURLPath, http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     cookieNameJWTToken,
		Value:    jwtToken,
		Expires:  time.Now().Add(cookieExpireDuration),
		Path:     rootURLPath,
		HttpOnly: cookieHTTPOnly,
		Secure:   cookieSecure,
	})

	http.Redirect(writer, req, rootURLPath, http.StatusTemporaryRedirect)
}

func (that *authHandler) getUserDataFromGoogle(ctx context.Context, code string) (*entity.User, error) {
	// use code to get token and get user info from Google.
	token, err := that.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %w", err)
	}

	response, err := that.oauthConfig.Client(ctx, token).Get(googleUserInfoURL) //nolint: noctx // ctx is in the client
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer response.Body.Close()

	var googleUserInfo struct {
		Email string `json:"email"`
	}
	if err = json.NewDecoder(response.Body).Decode(&googleUserInfo); err != nil {
		return nil, fmt.Errorf("failed decoding user info: %w", err)
	}

	userInfo := &entity.User{
		Email: googleUserInfo.Email,
	}

	return userInfo, nil
}

func (that *authHandler) generateStateOauthCookie(w http.ResponseWriter) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     cookieNameOAuthState,
		Value:    state,
		Expires:  time.Now().Add(oauthCookieExpire),
		HttpOnly: cookieHTTPOnly,
		Secure:   cookieSecure,
	}
	http.SetCookie(w, &cookie)

	return state, nil
}
