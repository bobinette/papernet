package jwt

import (
	"context"

	"github.com/dgrijalva/jwt-go"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
)

func Middleware(key []byte) endpoint.Middleware {
	claims := Claims{}
	return kitjwt.NewParser(func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.SigningMethodHS256, &claims)
}

func OptionalMiddleware(key []byte) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			// tokenString is stored in the context from the transport handlers.
			_, ok := ctx.Value(kitjwt.JWTTokenContextKey).(string)
			if !ok {
				return next(ctx, request)
			}

			return Middleware(key)(next)(ctx, request)
		}
	}
}
