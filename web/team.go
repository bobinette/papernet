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

func errTeamNotFound(id int) error {
	return errors.New(fmt.Sprintf("<Team %d> not found", id), errors.WithCode(http.StatusNotFound))
}

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
			HandlerFunc:   WrapRequest(h.create),
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
			HandlerFunc:   WrapRequest(h.get),
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
			Name:          "team.share",
			URL:           "/teams/:id/share",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.share),
		},
		papernet.EndPoint{
			Name:          "team.invite",
			URL:           "/teams/:id/invite",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.invite),
		},
		papernet.EndPoint{
			Name:          "team.kick",
			URL:           "/teams/:id/kick",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.kick),
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
	if !isStringIn(user.ID, team.Members) {
		return nil, errTeamNotFound(team.ID)
	}

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

	formatted, err := h.formatTeam(team)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": formatted,
	}, nil
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

	formattedTeams := make([]formattedTeam, len(teams))
	for i, team := range teams {
		formattedTeams[i], err = h.formatTeam(team)
		if err != nil {
			return nil, err
		}
	}
	return map[string]interface{}{
		"data": formattedTeams,
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

	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}

	// Check if user is a member. If not -> 404
	if !isStringIn(user.ID, team.Members) {
		return nil, errTeamNotFound(team.ID)
	}
	// Check if user is an admin. If not -> 403
	if !isStringIn(user.ID, team.Admins) {
		return nil, errors.New("only team admins can delete a team", errors.WithCode(http.StatusForbidden))
	}

	err = h.Store.Delete(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": "ok",
	}, nil
}

func (h *TeamHandler) get(req *Request) (interface{}, error) {
	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
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
	if !isStringIn(user.ID, team.Members) {
		return nil, errTeamNotFound(id)
	}

	formatted, err := h.formatTeam(team)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": formatted,
	}, nil
}

func (h *TeamHandler) create(req *Request) (interface{}, error) {
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

	formatted, err := h.formatTeam(team)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": formatted,
	}, nil
}

func (h *TeamHandler) invite(req *Request) (interface{}, error) {
	defer req.Body.Close()

	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
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

	if !isStringIn(user.ID, team.Admins) {
		return nil, errors.New("only team admins can invite new users", errors.WithCode(http.StatusForbidden))
	}

	body := struct {
		Email string `json:"email"`
	}{}
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	invidedUser, err := h.UserStore.Search(body.Email)
	if err != nil {
		return nil, err
	}
	if invidedUser == nil {
		return nil, errors.New(
			fmt.Sprintf("could not find user with email %s", body.Email),
			errors.WithCode(http.StatusNotFound),
		)
	}

	if !isStringIn(invidedUser.ID, team.Members) {
		team.Members = append(team.Members, invidedUser.ID)
		if err := h.Store.Upsert(&team); err != nil {
			return nil, err
		}
	}

	// Add papers to user
	for _, pID := range team.CanSee {
		if !isIn(pID, invidedUser.CanSee) {
			invidedUser.CanSee = append(invidedUser.CanSee, pID)
		}
	}

	for _, pID := range team.CanEdit {
		if !isIn(pID, invidedUser.CanEdit) {
			invidedUser.CanEdit = append(invidedUser.CanEdit, pID)
		}
	}

	formatted, err := h.formatTeam(team)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": formatted,
	}, nil
}

func (h *TeamHandler) kick(req *Request) (interface{}, error) {
	defer req.Body.Close()

	idStr := req.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("id should be an int", errors.WithCause(err), errors.WithCode(http.StatusBadRequest))
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

	if !isStringIn(user.ID, team.Admins) {
		return nil, errors.New("only team admins can kick members", errors.WithCode(http.StatusForbidden))
	}

	body := struct {
		UserID string `json:"userID"`
	}{}
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	if isStringIn(body.UserID, team.Members) {
		members := make([]string, len(team.Members)-1)
		offset := 0
		for i, memberID := range team.Members {
			if memberID == body.UserID {
				offset = 1
				continue
			}
			members[i-offset] = memberID
		}

		team.Members = members
	}

	if isStringIn(body.UserID, team.Admins) {
		admins := make([]string, len(team.Admins)-1)
		offset := 0
		for i, adminID := range team.Admins {
			if adminID == body.UserID {
				offset = 1
				continue
			}
			admins[i-offset] = adminID
		}

		team.Admins = admins
	}

	if err := h.Store.Upsert(&team); err != nil {
		return nil, err
	}

	formatted, err := h.formatTeam(team)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": formatted,
	}, nil
}

type teamMember struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Admin bool   `json:"admin"`
}

type formattedTeam struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	Members []teamMember `json:"members"`

	CanSee  []int `json:"canSee"`
	CanEdit []int `json:"canEdit"`
}

func (h *TeamHandler) formatTeam(team papernet.Team) (formattedTeam, error) {
	formatted := formattedTeam{
		ID:   team.ID,
		Name: team.Name,

		Members: make([]teamMember, len(team.Members)),

		CanSee:  team.CanSee,
		CanEdit: team.CanEdit,
	}

	for i, memberID := range team.Members {
		member, err := h.UserStore.Get(memberID)
		if err != nil {
			return formattedTeam{}, errors.New("error getting member")
		}

		formatted.Members[i] = teamMember{
			ID:    member.ID,
			Name:  member.Name,
			Email: member.Email,
			Admin: isStringIn(member.ID, team.Admins),
		}
	}

	return formatted, nil
}
