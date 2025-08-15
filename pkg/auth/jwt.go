package auth

import (
	"time"

	"github.com/flick/backend/services/user/proto"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey []byte

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

func GenerateJWT(user *proto.User, secret string) (string, error) {
	jwtKey = []byte(secret)
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   user.Id,
		Username: user.Username,
		Email:    user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	jwtKey = []byte(secret)
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}
