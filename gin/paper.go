package gin

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
)

type PaperHandler struct {
	Repository papernet.PaperRepository
	Searcher   papernet.PaperIndex

	UserRepository papernet.UserRepository
	Encoder        auth.Encoder
	Authenticator  Authenticator

	TagIndex papernet.TagIndex
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
		return nil, err
	}

	papers, err := h.Repository.Get(id)
	if err != nil {
		return nil, err
	} else if len(papers) == 0 {
		return nil, fmt.Errorf("Paper %d not found", id)
	}

	return map[string]interface{}{
		"data": papers[0],
	}, nil
}

func (h *PaperHandler) Insert(c *gin.Context) (interface{}, error) {
	var paper papernet.Paper
	err := c.BindJSON(&paper)
	if err != nil {
		return nil, err
	}

	if paper.ID > 0 {
		return nil, fmt.Errorf("field id should be empty, got %d", paper.ID)
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

func (h *PaperHandler) Update(c *gin.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, err
	}

	var paper papernet.Paper
	err = c.BindJSON(&paper)
	if err != nil {
		return nil, err
	}

	papersFromID, err := h.Repository.Get(id)
	if err != nil {
		return nil, err
	} else if len(papersFromID) == 0 {
		return nil, fmt.Errorf("Paper %d not found", id)
	}

	if paper.ID != id {
		return nil, fmt.Errorf("ids do not match: %d (url) != %d (body)", id, paper.ID)
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
		return nil, err
	}

	papers, err := h.Repository.Get(id)
	if err != nil {
		return nil, err
	} else if len(papers) == 0 {
		return nil, fmt.Errorf("Paper %d not found", id)
	}

	err = h.Repository.Delete(id)
	if err != nil {
		return nil, err
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
