package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthService interface {
	GenerateJWTToken(email string) (string, error)
	GenerateStateOauthSession(c echo.Context) (string, error)
}

type authService struct {
	secretKey string
}

func NewAuthService(secretKey string) AuthService {
	return &authService{
		secretKey: secretKey,
	}
}

func (that *authService) GenerateStateOauthSession(c echo.Context) (string, error) {
	// right now we're using generation via rand. Maybe we can use uuid to generate state.
	randBytes := make([]byte, 16)

	_, err := rand.Read(randBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	stateToken := base64.URLEncoding.EncodeToString(randBytes)

	// get user session.
	userSession, err := session.Get("user-session", c)
	if err != nil {
		return "", fmt.Errorf("failed to get user session: %v", err)
	}

	// save "state" to session.
	userSession.Values["state"] = stateToken
	if err = userSession.Save(c.Request(), c.Response()); err != nil {
		return "", fmt.Errorf("failed to save session: %v", err)
	}

	return stateToken, nil
}

func (that *authService) GenerateJWTToken(email string) (string, error) {
	claims := jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "tictactoe",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(that.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
