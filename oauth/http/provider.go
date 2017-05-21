package http

import (
	"context"
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
func RegisterProviderHTTPRoutes(srv Server, service *services.ProviderService) {
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

func makeProviderEndpoint(s *services.ProviderService) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		return map[string]interface{}{"providers": s.Providers()}, nil
	}
}

func decodeProviderRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}
