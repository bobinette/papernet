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
	Repository papernet.PaperRepository
	Searcher   papernet.PaperIndex

	TagIndex papernet.TagIndex

	Authenticator Authenticator
}

func (h *PaperHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/api/papers/:id", JSONFormatter(h.Get))
	router.PUT("/api/papers/:id", JSONFormatter(h.Update))
	router.DELETE("/api/papers/:id", JSONFormatter(h.Delete))
	router.GET("/api/papers", JSONFormatter(h.Authenticator.Authenticate(h.List)))
	router.POST("/api/papers", JSONFormatter(h.Insert))
}

func (h *PaperHandler) Get(c *gin.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	papers, err := h.Repository.Get(id)
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
	var paper papernet.Paper
	err := c.BindJSON(&paper)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	if paper.ID > 0 {
		return nil, errors.New(
			fmt.Sprintf("field id should be empty, got %d", paper.ID),
			errors.WithCode(http.StatusBadRequest),
		)
	}

	err = h.Repository.Upsert(&paper)
	if err != nil {
		return nil, errors.New("error inserting paper", errors.WithCause(err))
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	var paper papernet.Paper
	err = c.BindJSON(&paper)
	if err != nil {
		return nil, errors.New("error decoding json body", errors.WithCause(err))
	}

	papers, err := h.Repository.Get(id)
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

	err = h.Repository.Upsert(&paper)
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, errors.New("id should be an integer", errors.WithCode(http.StatusBadRequest))
	}

	papers, err := h.Repository.Get(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error retrieving paper %d", id),
			errors.WithCause(err),
		)
	} else if len(papers) == 0 {
		return nil, errors.New(fmt.Sprintf("Paper %d not found", id), errors.WithCode(http.StatusNotFound))
	}

	err = h.Repository.Delete(id)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("error deleting paper in the database %d", id),
			errors.WithCause(err),
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

	search := papernet.PaperSearch{
		Q: q,
	}

	if bookmarked {
		user, err := GetUser(c)
		if err != nil {
			return nil, err
		}
		search.IDs = user.Bookmarks
	}

	ids, err := h.Searcher.Search(search)
	if err != nil {
		return nil, err
	}

	papers, err := h.Repository.Get(ids...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": papers,
	}, nil
}
