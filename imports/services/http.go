package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/jwt"

	"github.com/bobinette/papernet/imports"
)

var (
	errInvalidRequest = errors.New("invalid request")
	errNoUser         = errors.New("no user", errors.WithCode(http.StatusUnauthorized))
)

// Server defines the interface to register the http handlers.
type HTTPServer interface {
	RegisterHandler(path, method string, f http.Handler)
}

// encodeError writes an error as an HTTP response. It handles the status code
// contained in the error.
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

// extractUserID returns the user id present in the context, or an error if there is
// no user id or the claims are not correct.
func extractUserID(ctx context.Context) (int, error) {
	claims := ctx.Value(kitjwt.JWTClaimsContextKey)
	if claims == nil {
		return 0, errNoUser
	}

	ppnClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return 0, errors.New("invalid claims", errors.WithCode(http.StatusForbidden))
	}

	return ppnClaims.UserID, nil
}

func (s *ImportService) RegisterHTTP(srv HTTPServer, jwtKey []byte) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}
	authenticationMiddleware := jwt.Middleware(jwtKey)

	searchHandler := kithttp.NewServer(
		makeSearchEndpoint(s),
		decodeSearchRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	sourcesHandler := kithttp.NewServer(
		authenticationMiddleware(makeSourcesEndpoint(s)),
		decodeSourcesRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	importHandler := kithttp.NewServer(
		authenticationMiddleware(makeImportEndpoint(s)),
		decodeImportRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	srv.RegisterHandler("/imports/v2/search", "GET", searchHandler)
	srv.RegisterHandler("/imports/v2/sources", "GET", sourcesHandler)
	srv.RegisterHandler("/imports/v2/import", "POST", importHandler)
}

func makeSourcesEndpoint(s *ImportService) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		return map[string]interface{}{
			"sources": s.Sources(),
		}, nil
	}
}

func decodeSourcesRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

type searchRequest struct {
	q       string
	limit   int
	offset  int
	sources []string
}

func makeSearchEndpoint(s *ImportService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		req, ok := r.(searchRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		userID, err := extractUserID(ctx)
		if err != nil && err != errNoUser {
			return nil, err
		}

		return s.Search(userID, req.q, req.limit, req.offset, req.sources, ctx)
	}
}

func decodeSearchRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	q := r.URL.Query().Get("q")

	limit := 0
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, errors.New("error reading limit parameter", errors.WithCause(err), errors.BadRequest())
		}
		limit = l
	}
	if limit <= 0 {
		limit = 20
	}

	offset := 0
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, errors.New("error reading offset parameter", errors.WithCause(err), errors.BadRequest())
		}
		offset = o
	}

	sources := r.URL.Query()["sources"]

	return searchRequest{
		q:       q,
		limit:   limit,
		offset:  offset,
		sources: sources,
	}, nil
}

func makeImportEndpoint(s *ImportService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		req, ok := r.(imports.Paper)
		if !ok {
			return nil, errInvalidRequest
		}

		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		return s.Import(userID, req, ctx)
	}
}

func decodeImportRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var paper imports.Paper
	err := json.NewDecoder(r.Body).Decode(&paper)
	if err != nil {
		return nil, err
	}
	return paper, nil
}
