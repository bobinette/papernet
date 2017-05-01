package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/bobinette/papernet/jwt"
)

func RegisterTeamHTTP(srv HTTPServer, service *TeamService, jwtKey []byte) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
		kithttp.ServerBefore(kitjwt.ToHTTPContext()),
	}
	authenticationMiddleware := jwt.Middleware(jwtKey)

	// Create endpoint
	ep := NewTeamEndpoint(service)

	// User teams handler
	userTeamsHandler := kithttp.NewServer(
		authenticationMiddleware(ep.UserTeams),
		decodeUserTeamsRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Create team handler
	createTeamHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Create),
		decodeCreateTeamRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Invite user handler
	inviteHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Invite),
		decodeInviteRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Kick user handler
	kickHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Kick),
		decodeKickRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Share paper handler
	shareHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Share),
		decodeShareRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Delete team handler
	deleteHandler := kithttp.NewServer(
		authenticationMiddleware(ep.Delete),
		decodeDeleteTeamRequest,
		kithttp.EncodeJSONResponse,
		opts...,
	)

	// Register all handlers
	srv.RegisterHandler("/auth/v2/teams", "GET", userTeamsHandler)
	srv.RegisterHandler("/auth/v2/teams", "POST", createTeamHandler)
	srv.RegisterHandler("/auth/v2/teams/:id", "DELETE", deleteHandler)
	srv.RegisterHandler("/auth/v2/teams/:id/invite", "POST", inviteHandler)
	srv.RegisterHandler("/auth/v2/teams/:id/kick", "POST", kickHandler)
	srv.RegisterHandler("/auth/v2/teams/:id/share", "POST", shareHandler)
}

func decodeUserTeamsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body
	return nil, nil
}

func decodeCreateTeamRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close() // Close body

	// Decode team from body
	var req Team
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func decodeInviteRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	teamID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	var body struct {
		Email string `json:"email"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	req := inviteRequest{
		TeamID: teamID,
		Email:  body.Email,
	}
	return req, nil
}

func decodeKickRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	teamID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	var body struct {
		ID int `json:"userID"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	req := kickRequest{
		TeamID:   teamID,
		MemberID: body.ID,
	}
	return req, nil
}

func decodeShareRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	teamID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	var body struct {
		PaperID int  `json:"paperID"`
		CanEdit bool `json:"canEdit"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	req := shareRequest{
		TeamID:  teamID,
		PaperID: body.PaperID,
		CanEdit: body.CanEdit,
	}
	return req, nil
}

func decodeDeleteTeamRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	params := ctx.Value("params").(map[string]string)
	teamID, err := strconv.Atoi(params["id"])
	if err != nil {
		return nil, err
	}

	req := deleteTeamRequest{
		TeamID: teamID,
	}
	return req, nil
}
