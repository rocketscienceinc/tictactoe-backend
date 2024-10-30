package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/rocketscienceinc/tittactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
	"github.com/rocketscienceinc/tittactoe-backend/internal/pkg"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	urlAccessToken = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
)

type Handlers interface {
	PingHandler(w http.ResponseWriter, _ *http.Request)

	GoogleLogin(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
}

type userService interface {
	SaveUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

type authService interface {
	GenerateToken(email string) (string, error)
}

type handlers struct {
	oauthConfig      *oauth2.Config
	oauthStateString string
	userService      userService
	authService      authService
}

func NewHandlers(redirectURL, clientID, clientSecret string, userService userService, authService authService) Handlers {
	oauthConfig := &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/user.email"},
		Endpoint:     google.Endpoint,
	}

	return &handlers{
		oauthConfig:      oauthConfig,
		userService:      userService,
		authService:      authService,
		oauthStateString: pkg.GenerateNewSessionID(),
	}
}

func (that *handlers) PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (that *handlers) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := pkg.GenerateNewSessionID()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	url := that.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (that *handlers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		http.Error(w, "State cookie not found", http.StatusBadRequest)
		return
	}
	state := stateCookie.Value

	queryState := r.URL.Query().Get("state")
	if state != queryState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in request", http.StatusBadRequest)
		return
	}

	token, err := that.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Code exchange failed", http.StatusInternalServerError)
		return
	}

	client := that.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get(urlAccessToken + token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}

	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	user, err := that.userService.GetUserByEmail(ctx, userInfo.Email)
	if errors.Is(err, apperror.ErrNotSavedUser) {
		user = &entity.User{
			Email: userInfo.Email,
		}
		if err = that.userService.SaveUser(ctx, user); err != nil {
			http.Error(w, "Failed to save user", http.StatusInternalServerError)
			return
		}
	}

	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	tokenString, err := that.authService.GenerateToken(user.Email)
	if err != nil {
		http.Error(w, "Failed to generate auth token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
