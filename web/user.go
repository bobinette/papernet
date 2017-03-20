package web

import (
	"encoding/json"
	"sort"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/errors"
)

type UserHandler struct {
	GoogleClient *auth.GoogleClient
	Store        papernet.UserStore
	Encoder      auth.TokenEncoder
}

func (h *UserHandler) Routes() []papernet.EndPoint {
	return []papernet.EndPoint{
		papernet.EndPoint{
			URL:           "/auth",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: false,
			HandlerFunc:   WrapRequest(h.AuthURL),
		},
		papernet.EndPoint{
			URL:           "/auth/google",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: false,
			HandlerFunc:   WrapRequest(h.Google),
		},
		papernet.EndPoint{
			URL:           "/me",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Me),
		},
		papernet.EndPoint{
			URL:           "/bookmarks",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.UpdateBookmarks),
		},
	}
}

func (h *UserHandler) AuthURL(*Request) (interface{}, error) {
	return map[string]interface{}{
		"url": h.GoogleClient.LoginURL(),
	}, nil
}

func (h *UserHandler) Google(req *Request) (interface{}, error) {
	var state string
	err := req.Query("state", &state)
	if err != nil {
		return nil, err
	}

	var code string
	err = req.Query("code", &code)
	if err != nil {
		return nil, err
	}

	user, err := h.GoogleClient.ExchangeToken(state, code)
	if err != nil {
		return nil, errors.New("error exchanging token", errors.WithCause(err))
	}

	if dbUser, err := h.Store.Get(user.ID); err != nil {
		return nil, errors.New("error checking user in db", errors.WithCause(err))
	} else {
		dbUser.Name = user.Name
		dbUser.Email = user.Email

		user = dbUser
		err = h.Store.Upsert(user)
		if err != nil {
			return nil, errors.New("error saving user", errors.WithCause(err))
		}
	}

	token, err := h.Encoder.Encode(user.ID)
	if err != nil {
		return nil, errors.New("error encoding token", errors.WithCause(err))
	}

	return map[string]interface{}{
		"access_token": token,
	}, nil
}

func (h *UserHandler) Me(req *Request) (interface{}, error) {
	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": user,
	}, nil
}

func (h *UserHandler) UpdateBookmarks(req *Request) (interface{}, error) {
	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}

	var payload struct {
		Add    []int `json:"add"`
		Remove []int `json:"remove"`
	}
	body := req.Body
	defer body.Close()
	err = json.NewDecoder(body).Decode(&payload)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	bookmarks := make(map[int]struct{})
	for _, b := range user.Bookmarks {
		bookmarks[b] = struct{}{}
	}
	for _, b := range payload.Add {
		bookmarks[b] = struct{}{}
	}
	for _, b := range payload.Remove {
		delete(bookmarks, b)
	}

	user.Bookmarks = func(m map[int]struct{}) []int {
		i := 0
		a := make([]int, len(m))
		for k, _ := range m {
			a[i] = k
			i++
		}
		sort.Ints(a)
		return a
	}(bookmarks)

	err = h.Store.Upsert(user)
	if err != nil {
		return nil, errors.New("error saving", errors.WithCause(err))
	}

	return map[string]interface{}{
		"data": user,
	}, nil
}
