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

// Server defines the interface to register the http handlers.
type Server interface {
	RegisterHandler(path, method string, f http.Handler)
}

func RegisterProviderHTTPRoutes(srv Server, service *ProviderService) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	providerHandler := kithttp.NewServer(
		makeProviderEndpoint(service),
		decodeProviderRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/login/providers", "GET", providerHandler)
}

func makeProviderEndpoint(s *ProviderService) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		return map[string]interface{}{"providers": s.Providers()}, nil
	}
}

func decodeProviderRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
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
