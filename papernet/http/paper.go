package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/users"

	"github.com/bobinette/papernet/clients/auth"

	"github.com/bobinette/papernet/papernet"
	"github.com/bobinette/papernet/papernet/endpoints"
	"github.com/bobinette/papernet/papernet/services"
)

func RegisterPaperEndpoints(srv Server, service *services.PaperService, jwtKey []byte, au *auth.Client) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}

	authenticator := users.NewAuthenticator(au)
	jwtMiddleware := jwt.Middleware(jwtKey)

	// Create endpoint
	ep := endpoints.NewPaperEndpoint(service)

	// Search paper handler
	searchPaperHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(ep.Search)),
		decodeSearchPaperRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Create paper handler
	createPaperHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(ep.Create)),
		decodeCreatePaperRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Get paper handler
	getPaperHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(ep.Get)),
		decodeGetPaperRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Update paper handler
	updatePaperHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(ep.Update)),
		decodeUpdatePaperRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Update paper handler
	deletePaperHandler := kithttp.NewServer(
		jwtMiddleware(authenticator.Authenticated(ep.Delete)),
		decodeGetPaperRequest, // Decoder is the same as get
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Register all handlers
	srv.RegisterHandler("/paper/v2/papers", "GET", searchPaperHandler)
	srv.RegisterHandler("/paper/v2/papers", "POST", createPaperHandler)
	srv.RegisterHandler("/paper/v2/papers/:id", "GET", getPaperHandler)
	srv.RegisterHandler("/paper/v2/papers/:id", "PUT", updatePaperHandler)
	srv.RegisterHandler("/paper/v2/papers/:id", "DELETE", deletePaperHandler)
}

func decodeGetPaperRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	paperID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	req := paperID
	return req, nil
}

func decodeSearchPaperRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	req := endpoints.SearchPaperRequest{}
	req.Q = r.URL.Query().Get("q")
	req.Tags = r.URL.Query()["tags"]

	bookmarked := r.URL.Query().Get("bookmarked")
	if bookmarked != "" {
		var err error
		req.Bookmarked, err = strconv.ParseBool(bookmarked)
		if err != nil {
			return nil, errors.New("invalid parameter: bookmarked", errors.BadRequest(), errors.WithCause(err))
		}
	}

	limit := r.URL.Query().Get("limit")
	if limit != "" {
		var err error
		req.Limit, err = strconv.Atoi(limit)
		if err != nil {
			return nil, errors.New("invalid parameter: limit", errors.BadRequest(), errors.WithCause(err))
		}
	}

	offset := r.URL.Query().Get("offset")
	if offset != "" {
		var err error
		req.Offset, err = strconv.Atoi(offset)
		if err != nil {
			return nil, errors.New("invalid parameter: offset", errors.BadRequest(), errors.WithCause(err))
		}
	}

	return req, nil
}

func decodeCreatePaperRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var paper papernet.Paper
	err := json.NewDecoder(r.Body).Decode(&paper)
	if err != nil {
		return nil, err
	}

	req := paper
	return req, nil
}

func decodeUpdatePaperRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	paperID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	var paper papernet.Paper
	err = json.NewDecoder(r.Body).Decode(&paper)
	if err != nil {
		return nil, err
	}

	if paper.ID != 0 && paperID != paper.ID {
		return nil, errors.New("ids do not match between url and body", errors.BadRequest())
	}

	req := paper
	return req, nil
}
