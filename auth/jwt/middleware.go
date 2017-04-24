package jwt

import (
	"github.com/dgrijalva/jwt-go"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
)

func Middleware(key []byte) endpoint.Middleware {
	claims := papernetClaims{}
	return kitjwt.NewParser(func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.SigningMethodHS256, &claims)
}
