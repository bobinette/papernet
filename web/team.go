package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/errors"
)

type TeamHandler struct {
	Store      papernet.TeamStore
	PaperStore papernet.PaperStore
	UserStore  papernet.UserStore
}

func (h *TeamHandler) Routes() []papernet.EndPoint {
	return []papernet.EndPoint{
		papernet.EndPoint{
			Name:          "team.create",
			URL:           "/teams",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Create),
		},
		papernet.EndPoint{
			Name:          "team.list",
			URL:           "/teams",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.list),
		},
		papernet.EndPoint{
			Name:          "team.get",
			URL:           "/teams/:id",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Get),
		},
		papernet.EndPoint{
			Name:          "team.update",
			URL:           "/teams/:id",
			Method:        "PUT",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Update),
		},
		papernet.EndPoint{
			Name:          "team.delete",
			URL:           "/teams/:id",
			Method:        "DELETE",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.delete),
		},
		papernet.EndPoint{
			Name:          "team.shared",
			URL:           "/teams/:id/share",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.share),
		},
	}
}

func (h *TeamHandler) share(req *Request) (interface{}, error) {
	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
	}

	body := struct {
		ID      int  `json:"id"`
		CanEdit bool `json:"canEdit"`
	}{}
	defer req.Body.Close()
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return nil, errors.New("invalid body", errors.WithCode(http.StatusBadRequest), errors.WithCause(err))
	}

	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}
	team, err := h.Store.Get(id)
	if err != nil {
		return nil, err
	} else if team.ID == 0 { // default
		return nil, errors.New(fmt.Sprintf("<Team %d> not found", id), errors.WithCode(http.StatusNotFound))
	}

	// Check if user is a member. If not -> 404

	papers, err := h.PaperStore.Get(body.ID)
	if err != nil {
		return nil, err
	} else if len(papers) == 0 {
		return nil, paperNotFound(body.ID)
	}

	paper := papers[0]

	userCanSee := false
	for _, pID := range user.CanSee {
		if pID == paper.ID {
			userCanSee = true
			break
		}
	}
	if !userCanSee {
		return nil, paperNotFound(paper.ID)
	}

	userCanEdit := false
	for _, pID := range user.CanEdit {
		if pID == paper.ID {
			userCanEdit = true
			break
		}
	}
	if body.CanEdit && !userCanEdit {
		return nil, errors.New("cannot grant edition", errors.WithCode(http.StatusForbidden))
	}

	// Add paper to team if not already present
	if isIn(body.ID, team.CanEdit) {
		// Already present
		return team, nil
	}

	if isIn(body.ID, team.CanSee) && !body.CanEdit {
		// Already in can see, and edit not set
		return team, nil
	}

	if !isIn(body.ID, team.CanSee) {
		team.CanSee = append(team.CanSee, body.ID)
	}
	if body.CanEdit && !isIn(body.ID, team.CanEdit) {
		team.CanEdit = append(team.CanEdit, body.ID)
	}
	err = h.Store.Upsert(&team)
	if err != nil {
		return nil, err
	}

	for _, memberID := range team.Members {
		member, err := h.UserStore.Get(memberID)
		if err != nil {
			return nil, err
		}

		if !isIn(body.ID, member.CanSee) {
			member.CanSee = append(member.CanSee, body.ID)
		}

		if body.CanEdit && !isIn(body.ID, member.CanEdit) {
			member.CanEdit = append(member.CanEdit, body.ID)
		}

		err = h.UserStore.Upsert(member)
		if err != nil {
			return nil, err
		}
	}

	return team, nil
}

func (h *TeamHandler) list(req *Request) (interface{}, error) {
	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}

	teams, err := h.Store.List(user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": teams,
	}, nil
}

func (h *TeamHandler) delete(req *Request) (interface{}, error) {
	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
	}

	team, err := h.Store.Get(id)
	if err != nil {
		return nil, err
	}

	if team.ID == 0 { // default value
		return nil, errors.New(fmt.Sprintf("<Team %d> not found", id), errors.WithCode(http.StatusNotFound))
	}

	// Check if user is a member. If not -> 404
	// Check if user is an admin. If not -> 403

	err = h.Store.Delete(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": "ok",
	}, nil
}

func (h *TeamHandler) Get(req *Request) (interface{}, error) {
	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
	}

	team, err := h.Store.Get(id)
	if err != nil {
		return nil, err
	} else if team.ID == 0 { // default
		return nil, errors.New(fmt.Sprintf("<Team %d> not found", id), errors.WithCode(http.StatusNotFound))
	}

	// Check if user is a member. If not -> 404

	return map[string]interface{}{
		"data": team,
	}, nil
}

func (h *TeamHandler) Create(req *Request) (interface{}, error) {
	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}

	teamBody := struct {
		Name string `json:"name"`
	}{}
	body := req.Body
	defer body.Close()
	err = json.NewDecoder(body).Decode(&teamBody)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	team := papernet.Team{
		Name: teamBody.Name,

		Admins:  []string{user.ID},
		Members: []string{user.ID},

		CanSee:  []int{},
		CanEdit: []int{},
	}
	if err := h.Store.Upsert(&team); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": team,
	}, nil
}

func (h *TeamHandler) Update(req *Request) (interface{}, error) {
	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
	}

	team, err := h.Store.Get(id)
	if err != nil {
		return nil, err
	}

	// Check if user is a member. If not -> 404
	// Check if user is admin. If not -> 403

	body := req.Body
	defer body.Close()
	err = json.NewDecoder(body).Decode(&team)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	if err := h.Store.Upsert(&team); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": team,
	}, nil
}
