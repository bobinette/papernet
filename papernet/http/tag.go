package http

import (
	"context"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/users"

	"github.com/bobinette/papernet/papernet/endpoints"
	"github.com/bobinette/papernet/papernet/services"

	auth "github.com/bobinette/papernet/auth/services"
)

func RegisterTagEndpoints(srv Server, service *services.TagService, jwtKey []byte, us *auth.UserService) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	authenticator := users.NewAuthenticator(us)
	jwtMiddleware := jwt.Middleware(jwtKey)

	// Create endpoint
	ep := endpoints.NewTagEndpoint(service)

	// Get paper handler
	getPaperHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(ep.Search)),
		decodeSearchTagRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Register all handlers
	srv.RegisterHandler("/paper/v2/tags", "GET", getPaperHandler)
}

func decodeSearchTagRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	req := r.URL.Query().Get("q")
	return req, nil
}
