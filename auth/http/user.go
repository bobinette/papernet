package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/endpoints"
	"github.com/bobinette/papernet/auth/services"
	"github.com/bobinette/papernet/jwt"
)

func RegisterUserEndpoints(srv Server, service *services.UserService, jwtKey []byte) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	jwtMiddleware := jwt.Middleware(jwtKey)

	// Create endpoint
	ep := endpoints.NewUserEndpoint(service)

	// Me handler
	meHandler := kithttp.NewServer(
		jwtMiddleware(ep.Me),
		decodeMeRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	userHandler := kithttp.NewServer(
		jwtMiddleware(ep.User),
		decodeUserRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	userUpsertHandler := kithttp.NewServer(
		jwtMiddleware(ep.Upsert),
		decodeUserUpsertRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	signUpHandler := kithttp.NewServer(
		ep.SignUp,
		decodeEmailPasswordRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	loginHandler := kithttp.NewServer(
		ep.Login,
		decodeEmailPasswordRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	tokenHandler := kithttp.NewServer(
		jwtMiddleware(ep.Token),
		decodeTokenRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	createPaperHandler := kithttp.NewServer(
		jwtMiddleware(ep.CreatePaper),
		decodeCreatePaperRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	bookmarkHandler := kithttp.NewServer(
		jwtMiddleware(ep.Bookmark),
		decodeBookmarkRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Routes
	srv.RegisterHandler("/auth/v2/me", "GET", meHandler)
	srv.RegisterHandler("/auth/v2/users/:id", "GET", userHandler)
	srv.RegisterHandler("/auth/v2/users", "POST", userUpsertHandler)
	srv.RegisterHandler("/auth/v2/signup", "POST", signUpHandler)
	srv.RegisterHandler("/auth/v2/login", "POST", loginHandler)
	srv.RegisterHandler("/auth/v2/users/:id/token", "GET", tokenHandler)
	srv.RegisterHandler("/auth/v2/users/:id/papers", "POST", createPaperHandler)
	srv.RegisterHandler("/auth/v2/bookmarks", "POST", bookmarkHandler)
}

func decodeMeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body
	return nil, nil
}

func decodeUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body

	params := ctx.Value("params").(map[string]string)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	return userID, nil
}

func decodeUserUpsertRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var user auth.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func decodeEmailPasswordRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var req endpoints.EmailPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return req, err
}

func decodeTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body

	params := ctx.Value("params").(map[string]string)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	return userID, nil
}

func decodeCreatePaperRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body

	params := ctx.Value("params").(map[string]string)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	var body struct {
		PaperID int `json:"paperID"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	req := endpoints.PaperCreateRequest{
		UserID:  userID,
		PaperID: body.PaperID,
	}
	return req, nil
}

func decodeBookmarkRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var body struct {
		PaperID  int  `json:"paperID"`
		Bookmark bool `json:"bookmark"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	req := endpoints.BookmarkRequest{
		PaperID:  body.PaperID,
		Bookmark: body.Bookmark,
	}
	return req, nil
}
