package auth

import (
	"context"
)

type TeamEndpoint struct {
	service *TeamService
}

func NewTeamEndpoint(s *TeamService) TeamEndpoint {
	return TeamEndpoint{
		service: s,
	}
}

func (ep TeamEndpoint) UserTeams(ctx context.Context, _ interface{}) (interface{}, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	return ep.service.GetForUser(userID)
}

func (ep TeamEndpoint) Create(ctx context.Context, r interface{}) (interface{}, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	team, ok := r.(Team)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Insert(userID, team)
}

type inviteRequest struct {
	TeamID int
	Email  string
}

func (ep TeamEndpoint) Invite(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(inviteRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Invite(callerID, req.TeamID, req.Email)
}

type kickRequest struct {
	TeamID   int
	MemberID int
}

func (ep TeamEndpoint) Kick(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(kickRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Kick(callerID, req.TeamID, req.MemberID)
}

type shareRequest struct {
	TeamID  int
	PaperID int
	CanEdit bool
}

func (ep TeamEndpoint) Share(ctx context.Context, r interface{}) (interface{}, error) {
	callerID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	req, ok := r.(shareRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	return ep.service.Share(callerID, req.TeamID, req.PaperID, req.CanEdit)
}
