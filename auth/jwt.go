package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenEncoder interface {
	Encode(int) (string, error)
}

type TokenDecoder interface {
	Decode(string) (int, error)
}

type EncodeDecoder struct {
	Key string
}

type papernetClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func (e EncodeDecoder) Encode(userID int) (string, error) {
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

func (e EncodeDecoder) Decode(bearer string) (int, error) {
	claims := papernetClaims{}

	token, err := jwt.ParseWithClaims(bearer, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(e.Key), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*papernetClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("could not get claims")
}
