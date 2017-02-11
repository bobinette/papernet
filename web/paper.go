package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/errors"
)

type PaperHandler struct {
	Store papernet.PaperStore
	Index papernet.PaperIndex

	TagIndex papernet.TagIndex

	UserStore papernet.UserRepository
}

func (h *PaperHandler) Routes() []papernet.Route {
	return []papernet.Route{
		papernet.Route{
			Route:         "/api/papers",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.List),
		},
		papernet.Route{
			Route:         "/api/papers",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Insert),
		},
		papernet.Route{
			Route:         "/api/papers/:id",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Get),
		},
		papernet.Route{
			Route:         "/api/papers/:id",
			Method:        "PUT",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Update),
		},
		papernet.Route{
			Route:         "/api/papers/:id",
			Method:        "DELETE",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Delete),
		},
	}
}

func (h *PaperHandler) List(req *Request) (interface{}, error) {
	q := ""
	_ = req.Query("q", &q)

	bookmarked := false
	err := req.Query("bookmarked", &bookmarked)
	if err != nil {
		return nil, err
	}

	user, err := req.User()
	if err != nil {
		return nil, err
	}

	var limit uint64 = 20
	err = req.Query("limit", &limit)
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = 20
	}

	var offset uint64 = 0
	err = req.Query("offset", &offset)
	if err != nil {
		return nil, err
	}

	search := papernet.PaperSearch{
		Q:      q,
		IDs:    user.CanSee,
		Limit:  limit,
		Offset: offset,
	}

	if bookmarked {
		search.IDs = user.Bookmarks
	}

	res, err := h.Index.Search(search)
	if err != nil {
		return nil, err
	}

	papers, err := h.Store.Get(res.IDs...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data":       papers,
		"pagination": res.Pagination,
	}, nil
}

func (h *PaperHandler) Insert(req *Request) (interface{}, error) {
	user, err := req.User()
	if err != nil {
		return nil, err
	}

	var paper papernet.Paper
	body := req.Body
	defer body.Close()
	err = json.NewDecoder(body).Decode(&paper)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	if paper.ID > 0 {
		return nil, errors.New(
			fmt.Sprintf("field id should be empty, got %d", paper.ID),
			errors.WithCode(http.StatusBadRequest),
		)
	}

	err = h.Store.Upsert(&paper)
	if err != nil {
		return nil, errors.New("error inserting paper", errors.WithCause(err))
	}

	// Give ownership
	user.CanSee = append(user.CanSee, paper.ID)
	user.CanEdit = append(user.CanEdit, paper.ID)
	err = h.UserStore.Upsert(user)
	if err != nil {
		return nil, errors.New("error setting rights on user", errors.WithCause(err))
	}

	err = h.Index.Index(&paper)
	if err != nil {
		return nil, errors.New("error indexing paper", errors.WithCause(err))
	}

	for _, tag := range paper.Tags {
		err = h.TagIndex.Index(tag)
		if err != nil {
			log.Printf("error indexing tag %s: %v", tag, err)
		}
	}

	return map[string]interface{}{
		"data": paper,
	}, nil
}

func (h *PaperHandler) Get(req *Request) (interface{}, error) {
	id, err := strconv.Atoi(req.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	user, err := req.User()
	if err != nil {
		return nil, err
	}

	if !isIn(id, user.CanSee) {
		return nil, errors.New(fmt.Sprintf("Paper %d not found", id), errors.WithCode(http.StatusNotFound))
	}

	papers, err := h.Store.Get(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error retrieving paper %d", id),
			errors.WithCause(err),
		)
	} else if len(papers) == 0 {
		return nil, errors.New(fmt.Sprintf("Paper %d not found", id), errors.WithCode(http.StatusNotFound))
	}

	return map[string]interface{}{
		"data": papers[0],
	}, nil
}

func (h *PaperHandler) Update(req *Request) (interface{}, error) {
	user, err := req.User()
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(req.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	var paper papernet.Paper
	body := req.Body
	defer body.Close()
	err = json.NewDecoder(body).Decode(&paper)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	papers, err := h.Store.Get(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error retrieving paper %d", id),
			errors.WithCause(err),
		)
	} else if len(papers) == 0 {
		return nil, errors.New(fmt.Sprintf("Paper %d not found", id), errors.WithCode(http.StatusNotFound))
	}

	if paper.ID != id {
		return nil, errors.New(
			fmt.Sprintf("ids do not match: %d (url) != %d (body)", id, paper.ID),
			errors.WithCode(http.StatusBadRequest),
		)
	}

	// Now we can check the permissions
	if !isIn(id, user.CanEdit) {
		return nil, errors.New(
			fmt.Sprintf("You are not allowed to edit Paper %d", id),
			errors.WithCode(http.StatusForbidden),
		)
	}

	err = h.Store.Upsert(&paper)
	if err != nil {
		return nil, err
	}

	err = h.Index.Index(&paper)
	if err != nil {
		return nil, err
	}

	for _, tag := range paper.Tags {
		err = h.TagIndex.Index(tag)
		if err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"data": paper,
	}, nil
}

func (h *PaperHandler) Delete(req *Request) (interface{}, error) {
	user, err := req.User()
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(req.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	papers, err := h.Store.Get(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error retrieving paper %d", id),
			errors.WithCause(err),
		)
	} else if len(papers) == 0 {
		return nil, errors.New(fmt.Sprintf("Paper %d not found", id), errors.WithCode(http.StatusNotFound))
	}

	if !isIn(id, user.CanEdit) {
		return nil, errors.New(
			fmt.Sprintf("You are not allowed to delete Paper %d", id),
			errors.WithCode(http.StatusForbidden),
		)
	}

	err = h.Index.Delete(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error deleting paper from the index %d", id),
			errors.WithCause(err),
		)
	}

	err = h.Store.Delete(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error deleting paper in the database %d", id),
			errors.WithCause(err),
		)
	}

	return map[string]interface{}{
		"data": "ok",
	}, nil
}

// ------------------------------------------------------------------------------------------
// Helpers

func isIn(i int, a []int) bool {
	for _, e := range a {
		if e == i {
			return true
		}
	}
	return false
}
