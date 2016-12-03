package gin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

type PaperHandler struct {
	Repository papernet.PaperRepository
	Searcher   papernet.PaperIndex

	UserRepository papernet.UserRepository
	SigningKey     papernet.SigningKey

	TagIndex papernet.TagIndex
}

func (h *PaperHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/papernet/papers/:id", h.Get)
	router.PUT("/papernet/papers/:id", h.Update)
	router.DELETE("/papernet/papers/:id", h.Delete)
	router.GET("/papernet/papers", h.List)
	router.POST("/papernet/papers", h.Insert)
}

func (h *PaperHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	papers, err := h.Repository.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err,
		})
		return
	} else if len(papers) == 0 {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": fmt.Sprintf("Paper %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": papers[0],
	})
}

func (h *PaperHandler) Insert(c *gin.Context) {
	var paper papernet.Paper
	err := c.BindJSON(&paper)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if paper.ID > 0 {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "field id should be empty",
		})
		return
	}

	err = h.Repository.Upsert(&paper)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err = h.Searcher.Index(&paper)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, tag := range paper.Tags {
		err = h.TagIndex.Index(tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": paper,
	})
}

func (h *PaperHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	var paper papernet.Paper
	err = c.BindJSON(&paper)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	papersFromID, err := h.Repository.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err,
		})
		return
	} else if len(papersFromID) == 0 {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": fmt.Sprintf("Paper %d not found", id),
		})
		return
	}

	if paper.ID != id {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": fmt.Sprintf("ids do not match: %d (url) != %d (body)", id, paper.ID),
		})
		return
	}

	err = h.Repository.Upsert(&paper)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err = h.Searcher.Index(&paper)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, tag := range paper.Tags {
		err = h.TagIndex.Index(tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": paper,
	})
}

func (h *PaperHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	papers, err := h.Repository.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	} else if len(papers) == 0 {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": fmt.Sprintf("Paper %d not found", id),
		})
		return
	}

	err = h.Repository.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": "ok",
	})
}

func (h *PaperHandler) List(c *gin.Context) {
	q := c.Query("q")
	bookmarked, _, err := queryBool("bookmarked", c)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	search := papernet.PaperSearch{
		Q: q,
	}

	if bookmarked {
		authHeader, ok := c.Request.Header["Authorization"]
		if !ok || len(authHeader) != 1 {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": "No token found",
			})
			return
		}

		token := authHeader[0]
		if !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "No bearer",
			})
			return
		}

		token = token[len("Bearer "):]
		userID, err := decodeToken(h.SigningKey.Key, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		user, err := h.UserRepository.Get(userID)
		search.IDs = user.Bookmarks
	}

	ids, err := h.Searcher.Search(search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	papers, err := h.Repository.Get(ids...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": papers,
	})
}
