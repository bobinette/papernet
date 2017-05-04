package endpoints

import (
	"context"

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
