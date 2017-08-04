package endpoints

import (
	"context"
	"net/http"

	"github.com/bobinette/papernet/errors"

	"github.com/bobinette/papernet/papernet"
	"github.com/bobinette/papernet/papernet/services"
	"github.com/bobinette/papernet/users"
)

// Variables and functions for specific errors
var (
	errInvalidRequest = errors.New("invalid request")
)

type PaperEndpoint struct {
	service *services.PaperService
}

func NewPaperEndpoint(service *services.PaperService) *PaperEndpoint {
	return &PaperEndpoint{
		service: service,
	}
}

type SearchPaperRequest struct {
	Q          string
	Tags       []string
	Bookmarked bool
	Limit      int
	Offset     int
}

func (ep *PaperEndpoint) Search(ctx context.Context, r interface{}) (interface{}, error) {
	user, err := users.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(SearchPaperRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	res, err := ep.service.Search(user, req.Q, req.Tags, req.Bookmarked, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data":       res.Papers,
		"pagination": res.Pagination,
		"facets":     res.Facets,
	}, nil
}

func (ep *PaperEndpoint) Create(ctx context.Context, r interface{}) (interface{}, error) {
	user, err := users.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	paper, ok := r.(papernet.Paper)
	if !ok {
		return nil, errInvalidRequest
	}

	paper, err = ep.service.Create(user.ID, paper)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": paper,
	}, nil
}

func (ep *PaperEndpoint) Get(ctx context.Context, r interface{}) (interface{}, error) {
	user, err := users.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	id, ok := r.(int)
	if !ok {
		return nil, errInvalidRequest
	}

	paper, err := ep.service.Get(user, id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": paper,
	}, nil
}

func (ep *PaperEndpoint) Update(ctx context.Context, r interface{}) (interface{}, error) {
	user, err := users.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	paper, ok := r.(papernet.Paper)
	if !ok {
		return nil, errInvalidRequest
	}

	paper, err = ep.service.Update(user, paper)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": paper,
	}, nil
}

func (ep *PaperEndpoint) Delete(ctx context.Context, r interface{}) (interface{}, error) {
	user, err := users.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	id, ok := r.(int)
	if !ok {
		return nil, errInvalidRequest
	}

	err = ep.service.Delete(user, id)
	if err != nil {
		return nil, err
	}

	return statusCoder{code: http.StatusNoContent}, nil
}

// statusCoder is useful to return http responses with a status that is not 200 but is not
// an error either.
type statusCoder struct {
	code int
}

func (s statusCoder) StatusCode() int { return s.code }
