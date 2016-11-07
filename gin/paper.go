package gin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

type PaperHandler struct {
	Repository papernet.PaperRepository
}

func (h *PaperHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/papernet/papers/:id", h.Get)
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

	paper, err := h.Repository.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err,
		})
		return
	} else if paper == nil {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": fmt.Sprintf("Paper %d not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": paper,
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

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": paper,
	})
}
