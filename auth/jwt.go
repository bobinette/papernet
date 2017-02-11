package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenEncoder interface {
	Encode(string) (string, error)
}

type TokenDecoder interface {
	Decode(string) (string, error)
}

type EncodeDecoder struct {
	Key string
}

type papernetClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func (e EncodeDecoder) Encode(userID string) (string, error) {
	claims := papernetClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 2, 0).Unix(),
			Issuer:    "papernet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(e.Key))
}

func (e EncodeDecoder) Decode(bearer string) (string, error) {
	claims := papernetClaims{}

	token, err := jwt.ParseWithClaims(bearer, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(e.Key), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*papernetClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("could not get claims")
}
