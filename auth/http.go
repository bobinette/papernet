package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/errors"

	"github.com/bobinette/papernet/auth/jwt"
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
func RegisterHTTPRoutes(srv Server, service *UserService, jwtKey []byte) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	authenticationMiddleware := jwt.Middleware(jwtKey)

	meHandler := kithttp.NewServer(
		authenticationMiddleware(makeMeEndpoint(service)),
		decodeMeRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/auth/me", "GET", meHandler)

	getUserHandler := kithttp.NewServer(
		authenticationMiddleware(makeGetUserEndpoint(service)),
		decodeGetUserRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/auth/users/:id", "GET", getUserHandler)

	upsertUserHandler := kithttp.NewServer(
		authenticationMiddleware(makeUpsertUserEndpoint(service)),
		decodeUpsertRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/auth/users", "POST", upsertUserHandler)
}

func decodeMeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	return nil, nil
}

func decodeGetUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	return userID, nil
}

func decodeUpsertRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		GoogleID string `json:"googleID"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	user := User{
		Name:     body.Name,
		Email:    body.Email,
		GoogleID: body.GoogleID,

		IsAdmin: false, // Never insert admin via web
	}
	return user, nil
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
