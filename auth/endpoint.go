package auth

import (
	"context"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"

	"github.com/bobinette/papernet/auth/jwt"
	"github.com/bobinette/papernet/errors"
)

type statusCoder struct {
	code int
}

func (s statusCoder) StatusCode() int { return s.code }

var (
	errInvalidRequest = errors.New("invalid request")
)

// --------------------------------------------
// Get user endpoints: get and me

func makeMeEndpoint(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		return s.Get(userID)
	}
}

func makeGetUserEndpoint(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, ok := r.(int)
		if !ok {
			return nil, errInvalidRequest
		}

		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		caller, err := s.Get(callerID)
		if err != nil {
			return nil, err
		}

		if userID != callerID && !caller.IsAdmin {
			return nil, errors.New("persmission denied, admin route", errors.WithCode(http.StatusForbidden))
		}

		return s.Get(userID)
	}
}

// --------------------------------------------
// Upsert user endpoint

func makeUpsertUserEndpoint(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		caller, err := s.Get(callerID)
		if err != nil {
			return nil, err
		}

		if !caller.IsAdmin {
			return nil, errors.New("persmission denied, admin route", errors.WithCode(http.StatusForbidden))
		}

		req, ok := r.(User)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.Upsert(req)
	}
}

// --------------------------------------------
// User -> paper endpoints

type bookmarkRequest struct {
	PaperID  int  `json:"paper_id"`
	Bookmark bool `json:"bookmark"`
}

func makeBookmarksHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(bookmarkRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.BookmarkPaper(userID, req.PaperID, req.Bookmark)
	}
}

type updateUserPapersRequest struct {
	PaperID int  `json:"paper_id"`
	Owns    bool `json:"owns"`

	UserID int
}

func makeUpdateUserPapersHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(updateUserPapersRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		caller, err := s.Get(callerID)
		if err != nil {
			return nil, err
		}

		if req.UserID != callerID && !caller.IsAdmin {
			return nil, errors.New("persmission denied, admin route", errors.WithCode(http.StatusForbidden))
		}

		return s.UpdateUserPapers(req.UserID, req.PaperID, req.Owns)
	}
}

// --------------------------------------------
// Teams endpoints

func makeMyTeamsHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		return s.UserTeams(userID)
	}
}

func makeInsertTeamHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(Team)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.InsertTeam(userID, req)
	}
}
func makeDeleteTeamHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		teamID, ok := r.(int)
		if !ok {
			return nil, errInvalidRequest
		}

		err = s.DeleteTeam(userID, teamID)
		if err != nil {
			return nil, err
		}
		return statusCoder{code: http.StatusNoContent}, nil
	}
}

type sharePaperRequest struct {
	TeamID  int
	PaperID int  `json:"paperID"`
	CanSee  bool `json:"canSee"`
	CanEdit bool `json:"canEdit"`
}

func makeSharePaperHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(sharePaperRequest)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.SharePaper(userID, req.TeamID, req.PaperID, req.CanSee, req.CanEdit)
	}
}

type inviteTeamMember struct {
	TeamID int
	Email  string `json:"email"`
	Admin  bool   `json:"admin"`
}

func makeInviteTeamMemberHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(inviteTeamMember)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.UpdateTeamMember(callerID, req.Email, req.TeamID, true, req.Admin)
	}
}

type kickTeamMember struct {
	TeamID int
	Email  string `json:"email"`
}

func makeKickTeamMemberHandler(s *UserService) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		callerID, err := extractUserID(ctx)
		if err != nil {
			return nil, err
		}

		req, ok := r.(kickTeamMember)
		if !ok {
			return nil, errInvalidRequest
		}

		return s.UpdateTeamMember(callerID, req.Email, req.TeamID, false, false)
	}
}

// --------------------------------------------
// Helpers

func extractUserID(ctx context.Context) (int, error) {
	claims := ctx.Value(kitjwt.JWTClaimsContextKey)
	if claims == nil {
		return 0, errors.New("no user", errors.WithCode(http.StatusUnauthorized))
	}

	ppnClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return 0, errors.New("invalid claims", errors.WithCode(http.StatusForbidden))
	}

	return ppnClaims.UserID, nil
}
