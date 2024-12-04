package service

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	jwtExpirationDuration = 24 * time.Hour
	tokenExpire           = "exp"
	tokenIat              = "iat"
)

type AuthService interface {
	GenerateJWT(userID string) (string, error)
}

type authService struct {
	jwtSecretKey string
}

func NewAuthService(jwtSecretKey string) AuthService {
	return &authService{
		jwtSecretKey: jwtSecretKey,
	}
}

func (that *authService) GenerateJWT(userEmail string) (string, error) {
	claims := jwt.StandardClaims{
		Subject:   userEmail,
		ExpiresAt: time.Now().Add(jwtExpirationDuration).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	decodedKey, err := base64.StdEncoding.DecodeString(that.jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT secret key: %w", err)
	}

	tokenString, err := token.SignedString(decodedKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}
