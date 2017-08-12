package endpoints

import (
	"context"
	"net/http"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/services"
)

type TeamEndpoint struct {
	service *services.TeamService
}

func NewTeamEndpoint(s *services.TeamService) TeamEndpoint {
	return TeamEndpoint{
		service: s,
	}
}

func (ep TeamEndpoint) UserTeams(ctx context.Context, _ interface{}) (interface{}, error) {
	userID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	return ep.service.GetForUser(userID)
}

func (ep TeamEndpoint) Create(ctx context.Context, r interface{}) (interface{}, error) {
	userID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	team, ok := r.(auth.Team)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Create(userID, team)
}

type InviteRequest struct {
	TeamID int
	Email  string
}

func (ep TeamEndpoint) Invite(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(InviteRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Invite(callerID, req.TeamID, req.Email)
}

type KickRequest struct {
	TeamID   int
	MemberID int
}

func (ep TeamEndpoint) Kick(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(KickRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Kick(callerID, req.TeamID, req.MemberID)
}

type ShareRequest struct {
	TeamID  int
	PaperID int
	CanEdit bool
}

func (ep TeamEndpoint) Share(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(ShareRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Share(callerID, req.TeamID, req.PaperID, req.CanEdit)
}

type DeleteTeamRequest struct {
	TeamID int
}

func (ep TeamEndpoint) Delete(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, _, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(DeleteTeamRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	err = ep.service.Delete(callerID, req.TeamID)
	if err != nil {
		return nil, err
	}
	return statusCoder{code: http.StatusNoContent}, nil
}
