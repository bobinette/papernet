package gin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/errors"
)

type PaperHandler struct {
	Store    papernet.PaperStore
	Searcher papernet.PaperIndex

	UserRepository papernet.UserRepository

	TagIndex papernet.TagIndex

	Authenticator Authenticator
}

func (h *PaperHandler) RegisterRoutes(router *gin.Engine) {
	router.PUT("/api/papers/:id", JSONFormatter(h.Authenticator.Authenticate(h.Update)))
	router.DELETE("/api/papers/:id", JSONFormatter(h.Authenticator.Authenticate(h.Delete)))
}

func (h *PaperHandler) Update(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	var paper papernet.Paper
	err = c.BindJSON(&paper)
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

	err = h.Searcher.Index(&paper)
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

func (h *PaperHandler) Delete(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(c.Param("id"))
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

	err = h.Store.Delete(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error deleting paper in the database %d", id),
			errors.WithCause(err),
		)
	}

	if !isIn(id, user.CanEdit) {
		return nil, errors.New(
			fmt.Sprintf("You are not allowed to delete Paper %d", id),
			errors.WithCode(http.StatusForbidden),
		)
	}

	err = h.Searcher.Delete(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error deleting paper from the index %d", id),
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
