package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bobinette/papernet/errors"
)

var (
	errInvalidRequest = errors.New("invalid request")
)

// Server defines the interface to register the http handlers.
type Server interface {
	RegisterHandler(path, method string, f http.Handler)
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
