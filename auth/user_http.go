package auth

import (
	"context"
	"encoding/json"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/jwt"
)

func RegisterUserHTTP(srv HTTPServer, service *UserService, jwtKey []byte) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}
	authenticationMiddleware := jwt.Middleware(jwtKey)

	// Create endpoint
	ep := NewUserEndpoint(service)

	// Me handler
	meHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Me),
		decodeMeRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	bookmarkHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Bookmark),
		decodeBookmarkRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Routes
	srv.RegisterHandler("/auth/v2/me", "GET", meHandler)
	srv.RegisterHandler("/auth/v2/bookmarks", "POST", bookmarkHandler)
}

func decodeMeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body
	return nil, nil
}

func decodeBookmarkRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var body struct {
		PaperID  int  `json:"paperID"`
		Bookmark bool `json"bookmark"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	req := bookmarkRequest{
		PaperID:  body.PaperID,
		Bookmark: body.Bookmark,
	}
	return req, nil
}
