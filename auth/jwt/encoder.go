package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Encoder struct {
	key []byte
}

type papernetClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func NewEncode(key []byte) *Encoder {
	return &Encoder{
		key: key,
	}
}

func (e *Encoder) Encode(userID int) (string, error) {
	claims := papernetClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 2, 0).Unix(),
			Issuer:    "papernet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(e.key)
}
