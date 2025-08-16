package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

func InitJWT(secret string) { jwtKey = []byte(secret) }

func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(jwtKey)
}

func ParseToken(tok string) (uint, error) {
	if len(jwtKey) == 0 {
		return 0, errors.New("jwt not initialized")
	}
	token, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if v, ok := claims["user_id"].(float64); ok { // JSON numbers are float64
			return uint(v), nil
		}
	}
	return 0, errors.New("user_id not in token")
}