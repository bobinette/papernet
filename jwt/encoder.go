package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/bobinette/papernet/errors"
)

type EncodeDecoder struct {
	key []byte
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func NewEncodeDecoder(key []byte) *EncodeDecoder {
	return &EncodeDecoder{
		key: key,
	}
}

func (e *EncodeDecoder) Encode(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 2, 0).Unix(),
			Issuer:    "papernet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(e.key)
}

func (e *EncodeDecoder) Decode(bearer string) (int, error) {
	claims := Claims{}

	token, err := jwt.ParseWithClaims(bearer, &claims, func(token *jwt.Token) (interface{}, error) {
		return e.key, nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("could not get claims")
}
