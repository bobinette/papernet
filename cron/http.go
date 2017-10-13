package cron

import (
	"context"
	"encoding/json"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/users"
)

var (
	errInvalidRequest = errors.New("invalid request")
	errNoUser         = errors.New("no user", errors.WithCode(http.StatusUnauthorized))
)

// Server defines the interface to register the http handlers.
type HTTPServer interface {
	RegisterHandler(path, method string, f http.Handler)
}

// encodeError writes an error as an HTTP response. It handles the status code
// contained in the error.
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	statusCode := http.StatusInternalServerError
	if err, ok := err.(errors.Error); ok {
		statusCode = err.Code()
	}
	w.WriteHeader(statusCode)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// extractUserID returns the user id present in the context, or an error if there is
// no user id or the claims are not correct.
func extractUserID(ctx context.Context) (int, error) {
	claims := ctx.Value(kitjwt.JWTClaimsContextKey)
	if claims == nil {
		return 0, errNoUser
	}

	ppnClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return 0, errors.New("invalid claims", errors.WithCode(http.StatusForbidden))
	}

	return ppnClaims.UserID, nil
}

func (s *Service) RegisterHTTP(srv HTTPServer, jwtKey []byte, authClient *auth.Client) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	authenticator := users.NewAuthenticator(authClient)
	authenticationMiddleware := jwt.Middleware(jwtKey)

	cronListHandler := kithttp.NewServer(
		authenticationMiddleware(authenticator.Valid(makeCronListEndpoint(s))),
		decodeCronListRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	cronInsertHandler := kithttp.NewServer(
		authenticationMiddleware(authenticator.Valid(makeCronInsertEndpoint(s))),
		decodeCronInsertRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	cronRunHandler := kithttp.NewServer(
		authenticationMiddleware(authenticator.Admin(makeCronRunEndpoint(s))),
		decodeCronRunRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/cron/v1/crons", "GET", cronListHandler)
	srv.RegisterHandler("/cron/v1/crons", "POST", cronInsertHandler)
	srv.RegisterHandler("/cron/v1/crons/run", "POST", cronRunHandler)
}

func makeCronListEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		crons, err := s.GetForUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"data": crons,
		}, nil
	}
}

func decodeCronListRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func makeCronInsertEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		cron, ok := r.(Cron)
		if !ok {
			return nil, errInvalidRequest
		}

		cron.UserID = userID
		err = s.Insert(ctx, &cron)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"data": cron,
		}, nil
	}
}

func decodeCronInsertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var req Cron
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func makeCronRunEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		err := s.RunCrons(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"data": "ok",
		}, nil
	}
}

func decodeCronRunRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}
