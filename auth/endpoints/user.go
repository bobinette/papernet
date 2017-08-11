package endpoints

import (
	"context"

	"github.com/bobinette/papernet/errors"

	"github.com/bobinette/papernet/auth/services"
)

type UserEndpoint struct {
	service *services.UserService
}

func NewUserEndpoint(s *services.UserService) UserEndpoint {
	return UserEndpoint{
		service: s,
	}
}

func (ep UserEndpoint) Me(ctx context.Context, _ interface{}) (interface{}, error) {
	callerID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	return ep.service.Get(callerID)
}

func (ep UserEndpoint) User(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	if caller, err := ep.service.Get(callerID); err != nil {
		return nil, err
	} else if !caller.IsAdmin {
		return nil, errors.New("admin route", errors.Forbidden())
	}

	userID, ok := r.(int)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Get(userID)
}

type BookmarkRequest struct {
	PaperID  int
	Bookmark bool
}

func (ep UserEndpoint) Bookmark(ctx context.Context, r interface{}) (interface{}, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(BookmarkRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Bookmark(userID, req.PaperID, req.Bookmark)
}
