package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/errors"
)

var (
	errInvalidRequest = errors.New("invalid request")
)

type Server interface {
	RegisterHandler(path, method string, f http.Handler)
}

// MakeHTTPHandler returns a http handler for the auth service. It defines the following routes:
// set/get paper owner,
// get user permissions,
// get user teams,
// add/remove user from a team,
// add/remove/get team permissions on a paper
func RegisterHTTPRoutes(srv Server, service *GoogleService) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	googleLoginURLHandler := kithttp.NewServer(
		makeGoogleLoginURLEndpoint(service),
		decodeGoogleLoginURLRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/login/google", "GET", googleLoginURLHandler)

	googleLoginHandler := kithttp.NewServer(
		makeGoogleLoginEndpoint(service),
		decodeGoogleLoginRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/login/google", "POST", googleLoginHandler)
}

func makeGoogleLoginURLEndpoint(s *GoogleService) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		return map[string]string{"url": s.LoginURL()}, nil
	}
}

func decodeGoogleLoginURLRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

type GoogleLoginRequest struct {
	State string `json:"state"`
	Code  string `json:"code"`
}

func makeGoogleLoginEndpoint(s *GoogleService) endpoint.Endpoint {
	return func(_ context.Context, r interface{}) (interface{}, error) {
		req, ok := r.(GoogleLoginRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.Login(req.State, req.Code)
	}
}

func decodeGoogleLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

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
