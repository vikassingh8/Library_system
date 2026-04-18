package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)




func GenerateJWT(userId int, role string, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userId,
		"role": role,
		"exp":  time.Now().Add(expiry).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyJWT validates the token and returns (userId, role, error).
func VerifyJWT(token string, secret string) (int, string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, "", err
	}
	userId := int(claims["sub"].(float64))
	role, _ := claims["role"].(string)
	return userId, role, nil
}