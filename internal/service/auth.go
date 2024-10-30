package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type AuthService interface {
	GenerateToken(email string) (string, error)
}

type authServiceImpl struct {
	secretKey string
}

func NewAuthService(secretKey string) AuthService {
	return &authServiceImpl{
		secretKey: secretKey,
	}
}

func (that *authServiceImpl) GenerateToken(email string) (string, error) {
	claims := jwt.MapClaims{}
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(that.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
