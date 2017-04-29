package auth

import (
	"context"
)

type UserEndpoint struct {
	service *UserService
}

func NewUserEndpoint(s *UserService) UserEndpoint {
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

type bookmarkRequest struct {
	PaperID  int
	Bookmark bool
}

func (ep UserEndpoint) Bookmark(ctx context.Context, r interface{}) (interface{}, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(bookmarkRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Bookmark(userID, req.PaperID, req.Bookmark)
}
