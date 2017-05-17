package endpoints

import (
	"context"

	"github.com/bobinette/papernet/papernet/services"
)

type TagEndpoint struct {
	service *services.TagService
}

func NewTagEndpoint(service *services.TagService) *TagEndpoint {
	return &TagEndpoint{
		service: service,
	}
}

func (ep *TagEndpoint) Search(ctx context.Context, r interface{}) (interface{}, error) {
	q, ok := r.(string)
	if !ok {
		return nil, errInvalidRequest
	}
	return ep.service.Search(q)
}
