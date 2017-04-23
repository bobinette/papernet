package auth

import (
	"context"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"

	"github.com/bobinette/papernet/errors"
)

var (
	errInvalidRequest = errors.New("invalid request")
)

// --------------------------------------------
// Get user endpoints: get and me

func makeMeEndpoint(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		return s.Get(userID)
	}
}

func makeGetUserEndpoint(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, ok := r.(int)
		if !ok {
			return nil, errInvalidRequest
		}

		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		user, err := s.Get(callerID)
		if err != nil {
			return nil, err
		}

		if userID != callerID && !user.IsAdmin {
			return nil, errors.New("persmission denied, admin route", errors.WithCode(http.StatusForbidden))
		}

		return s.Get(userID)
	}
}

// --------------------------------------------
// Upsert user endpoint

func makeUpsertUserEndpoint(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		user, err := s.Get(callerID)
		if err != nil {
			return nil, err
		}

		if !user.IsAdmin {
			return nil, errors.New("persmission denied, admin route", errors.WithCode(http.StatusForbidden))
		}

		req, ok := r.(User)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.Upsert(req)
	}
}

// --------------------------------------------
// Helpers

func extractUserID(ctx context.Context) (int, error) {
	claims := ctx.Value(kitjwt.JWTClaimsContextKey)
	if claims == nil {
		return 0, errors.New("no user", errors.WithCode(http.StatusUnauthorized))
	}

	ppnClaims, ok := claims.(*papernetClaims)
	if !ok {
		return 0, errors.New("invalid claims", errors.WithCode(http.StatusForbidden))
	}

	return ppnClaims.UserID, nil
}
