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
	callerID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	return ep.service.Get(callerID)
}

func (ep UserEndpoint) User(ctx context.Context, r interface{}) (interface{}, error) {
	_, isAdmin, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil, errors.New("admin route", errors.Forbidden())
	}

	userID, ok := r.(int)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Get(userID)
}

func (ep UserEndpoint) Token(ctx context.Context, r interface{}) (interface{}, error) {
	_, isAdmin, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	} else if !isAdmin {
		return nil, errors.New("admin route", errors.Forbidden())
	}

	userID, ok := r.(int)
	if !ok {
		return nil, errInvalidRequest
	}

	token, err := ep.service.Token(userID)
	if err != nil {
		return nil, err
	}
	return map[string]string{"access_token": token}, nil
}

type PaperCreateRequest struct {
	UserID  int
	PaperID int
}

func (ep UserEndpoint) CreatePaper(ctx context.Context, r interface{}) (interface{}, error) {
	_, isAdmin, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	} else if !isAdmin {
		return nil, errors.New("admin route", errors.Forbidden())
	}

	req, ok := r.(PaperCreateRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.CreatePaper(req.UserID, req.PaperID)
}

type BookmarkRequest struct {
	PaperID  int
	Bookmark bool
}

func (ep UserEndpoint) Bookmark(ctx context.Context, r interface{}) (interface{}, error) {
	userID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(BookmarkRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Bookmark(userID, req.PaperID, req.Bookmark)
}
