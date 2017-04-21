package auth

import (
	"context"
	"encoding/json"
	// "fmt"
	"net/http"
	// "strconv"

	"github.com/bobinette/papernet/errors"
	kithttp "github.com/go-kit/kit/transport/http"
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
func RegisterHTTPRoutes(srv Server, service *Service) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	meHandler := kithttp.NewServer(
		service.Get,
		decodeMeRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/auth/me", "GET", meHandler)

	upsertUserHandler := kithttp.NewServer(
		service.Upsert,
		decodeUpsertRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)
	srv.RegisterHandler("/auth/users", "POST", upsertUserHandler)
}

func decodeMeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	req := GetRequest{
		ID: 1, // @TODO: get from token
	}
	return req, nil
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

	req := UpsertRequest{
		Name:     body.Name,
		Email:    body.Email,
		GoogleID: body.GoogleID,
	}
	return req, nil
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
