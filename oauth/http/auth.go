package http

import (
	"context"
	"encoding/json"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/oauth/services"
)

// MakeHTTPHandler returns a http handler for the auth service. It defines the following routes:
// set/get paper owner,
// get user permissions,
// get user teams,
// add/remove user from a team,
// add/remove/get team permissions on a paper
func RegisterAuthHTTPRoutes(srv Server, service *services.AuthService) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	signUpHandler := kithttp.NewServer(
		makeSignUpEndpoint(service),
		decodeSignUpRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	loginHandler := kithttp.NewServer(
		makeLoginEndpoint(service),
		decodeLoginRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/signup", "POST", signUpHandler)
	srv.RegisterHandler("/login", "POST", loginHandler)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func makeSignUpEndpoint(s *services.AuthService) endpoint.Endpoint {
	return func(_ context.Context, r interface{}) (interface{}, error) {
		req, ok := r.(LoginRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		token, err := s.SignUp(req.Email, req.Password)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"access_token": token}, nil
	}
}

func decodeSignUpRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

func makeLoginEndpoint(s *services.AuthService) endpoint.Endpoint {
	return func(_ context.Context, r interface{}) (interface{}, error) {
		req, ok := r.(LoginRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		token, err := s.Login(req.Email, req.Password)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"access_token": token}, nil
	}
}

func decodeLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}
