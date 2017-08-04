package users

import (
	"context"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/jwt"

	"github.com/bobinette/papernet/auth/services"
)

var (
	contextKey = "user"
)

type User struct {
	ID      int
	IsAdmin bool

	Owns      []int
	CanSee    []int
	CanEdit   []int
	Bookmarks []int
}

func FromContext(ctx context.Context) (User, error) {
	v := ctx.Value(contextKey)
	if v == nil {
		return User{}, errors.New("no user", errors.WithCode(http.StatusUnauthorized))
	}

	user, ok := v.(User)
	if !ok {
		return User{}, errors.New("invalid user", errors.WithCode(http.StatusUnauthorized))
	}

	return user, nil
}

func extractUserID(ctx context.Context) (int, error) {
	claims := ctx.Value(kitjwt.JWTClaimsContextKey)
	if claims == nil {
		return 0, errors.New("no user", errors.WithCode(http.StatusUnauthorized))
	}

	ppnClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return 0, errors.New("invalid claims", errors.WithCode(http.StatusUnauthorized))
	}

	return ppnClaims.UserID, nil
}

type Authenticator struct {
	service *services.UserService
}

func NewAuthenticator(s *services.UserService) *Authenticator {
	return &Authenticator{
		service: s,
	}
}

func (a *Authenticator) get(id int) (User, error) {
	user, err := a.service.Get(id)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:      user.ID,
		IsAdmin: user.IsAdmin,

		Owns:      user.Owns,
		CanSee:    user.CanSee,
		CanEdit:   user.CanEdit,
		Bookmarks: user.Bookmarks,
	}, nil
}

func (a *Authenticator) Valid(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, contextKey, User{ID: userID})
		return next(ctx, req)
	}
}

func (a *Authenticator) Authenticated(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		user, err := a.get(userID)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, contextKey, user)
		return next(ctx, req)
	}
}

func (a *Authenticator) Admin(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		user, err := a.get(userID)
		if err != nil {
			return nil, err
		}

		if user.IsAdmin {
			return 0, errors.New("admin only", errors.WithCode(http.StatusForbidden))
		}

		ctx = context.WithValue(ctx, contextKey, user)
		return next(ctx, req)
	}
}
