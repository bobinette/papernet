package google

import (
	"context"
	"encoding/json"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/users"
)

func RegisterGoogleHTTPRoutes(srv Server, service *Service, jwtKey []byte, authClient *auth.Client) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	authenticator := users.NewAuthenticator(authClient)
	jwtMiddleware := jwt.Middleware(jwtKey)

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

	googleDriveRequireHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(makeGoogleDriveRequireHandler(service))),
		decodeGoogleDriveRequireRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	googleDriveViewHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(makeGoogleDriveViewHandler(service))),
		decodeGoogleDriveViewRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/google/login", "GET", googleLoginURLHandler)
	srv.RegisterHandler("/google/login", "POST", googleLoginHandler)
	srv.RegisterHandler("/google/drive/require", "GET", googleDriveRequireHandler)
	srv.RegisterHandler("/google/drive", "GET", googleDriveViewHandler)

	// @TODO: remove because legacy
	srv.RegisterHandler("/login/google", "GET", googleLoginURLHandler)
	srv.RegisterHandler("/login/google", "POST", googleLoginHandler)
}

func makeGoogleLoginURLEndpoint(s *Service) endpoint.Endpoint {
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

func makeGoogleLoginEndpoint(s *Service) endpoint.Endpoint {
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

func makeGoogleDriveViewHandler(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		user, err := users.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		err = s.inspectDrive(user.ID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func decodeGoogleDriveViewRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func makeGoogleDriveRequireHandler(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		user, err := users.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		hasAccess, url, err := s.requireDrive(user.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"url": url, "hasAccess": hasAccess}, nil
	}
}

func decodeGoogleDriveRequireRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}
