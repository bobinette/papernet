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

func RegisterGoogleHTTPRoutes(srv Server, service *services.GoogleService) {
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

	googleLoginHandler := kithttp.NewServer(
		makeGoogleLoginEndpoint(service),
		decodeGoogleLoginRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/login/google", "GET", googleLoginURLHandler)
	srv.RegisterHandler("/login/google", "POST", googleLoginHandler)
}

func makeGoogleLoginURLEndpoint(s *services.GoogleService) endpoint.Endpoint {
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

func makeGoogleLoginEndpoint(s *services.GoogleService) endpoint.Endpoint {
	return func(_ context.Context, r interface{}) (interface{}, error) {
		req, ok := r.(GoogleLoginRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		token, err := s.Login(req.State, req.Code)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"access_token": token}, nil
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
