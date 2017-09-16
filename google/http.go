package google

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

	googleDriveHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(makeGoogleDriveHandler(service))),
		decodeGoogleDriveRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	googleDriveRequireHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(makeGoogleDriveRequireHandler(service))),
		decodeGoogleDriveRequireRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	googleDriveFilesHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(makeGoogleDriveFilesHandler(service))),
		decodeGoogleDriveFilesRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	googleDriveUploadFileHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(makeGoogleDriveUploadFileHandler(service))),
		decodeGoogleDriveUploadFileRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/google/login", "GET", googleLoginURLHandler)
	srv.RegisterHandler("/google/login", "POST", googleLoginHandler)

	srv.RegisterHandler("/google/drive", "GET", googleDriveHandler)
	srv.RegisterHandler("/google/drive/require", "GET", googleDriveRequireHandler)
	srv.RegisterHandler("/google/drive/files", "GET", googleDriveFilesHandler)
	srv.RegisterHandler("/google/drive/files", "POST", googleDriveUploadFileHandler)

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

		token, fromURL, err := s.Login(req.State, req.Code)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"access_token": token, "fromURL": fromURL}, nil
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

func makeGoogleDriveHandler(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		user, err := users.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		has, err := s.hasDrive(user.ID)
		if err != nil {
			return nil, err
		}
		return map[string]bool{"hasAccess": has}, nil
	}
}

func decodeGoogleDriveRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func makeGoogleDriveRequireHandler(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		fromURL, ok := r.(string)
		if !ok {
			return nil, errInvalidRequest
		}

		url := s.requireDrive(fromURL)
		return map[string]interface{}{"url": url}, nil
	}
}

func decodeGoogleDriveRequireRequest(_ context.Context, r *http.Request) (interface{}, error) {
	fromURL := r.URL.Query().Get("fromURL")
	return fromURL, nil
}

type googleDriveFilesRequest struct {
	name string
}

func makeGoogleDriveFilesHandler(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		user, err := users.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(googleDriveFilesRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		files, pageToken, err := s.inspectDrive(user.ID, req.name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": files, "pageToken": pageToken}, nil
	}
}

func decodeGoogleDriveFilesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	name := r.URL.Query().Get("name")

	return googleDriveFilesRequest{
		name: name,
	}, nil
}

type googleDriveUploadRequest struct {
	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
	Data     []byte `json:"data"`
}

func makeGoogleDriveUploadFileHandler(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		user, err := users.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(googleDriveUploadRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		file, err := s.addFile(user.ID, req.Filename, req.Filetype, req.Data)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"file": file}, nil
	}
}

func decodeGoogleDriveUploadFileRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var body googleDriveUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, errors.New("error decoding body", errors.WithCause(err), errors.BadRequest())
	}

	if body.Filename == "" {
		return nil, errors.New("missing filename", errors.BadRequest())
	} else if body.Filetype == "" {
		return nil, errors.New("missing filetype", errors.BadRequest())
	}
	return body, nil
}
