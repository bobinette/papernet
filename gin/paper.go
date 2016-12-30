package gin

import (
	"fmt"
	"log"
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
	router.GET("/api/papers/:id", JSONFormatter(h.Authenticator.Authenticate(h.Get)))
	router.PUT("/api/papers/:id", JSONFormatter(h.Authenticator.Authenticate(h.Update)))
	router.DELETE("/api/papers/:id", JSONFormatter(h.Authenticator.Authenticate(h.Delete)))
	router.GET("/api/papers", JSONFormatter(h.Authenticator.Authenticate(h.List)))
	router.POST("/api/papers", JSONFormatter(h.Authenticator.Authenticate(h.Insert)))
}

func (h *PaperHandler) Get(c *gin.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	user, err := GetUser(c)
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

func (h *PaperHandler) Insert(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	var paper papernet.Paper
	err = c.BindJSON(&paper)
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
	err = h.UserRepository.Upsert(user)
	if err != nil {
		return nil, errors.New("error setting rights on user", errors.WithCause(err))
	}

	err = h.Searcher.Index(&paper)
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

func (h *PaperHandler) List(c *gin.Context) (interface{}, error) {
	q := c.Query("q")
	bookmarked, _, err := queryBool("bookmarked", c)
	if err != nil {
		return nil, err
	}

	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	limit, ok, err := queryInt("limit", c)
	if err != nil {
		return nil, err
	} else if !ok {
		limit = 20
	}

	offset, _, err := queryInt("offset", c)
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

	res, err := h.Searcher.Search(search)
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
