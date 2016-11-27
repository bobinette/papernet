package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

type TagHandler struct {
	Searcher papernet.TagIndex
}

func (h *TagHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/papernet/tags", h.List)
}

func (h *TagHandler) List(c *gin.Context) {
	q := c.Query("q")
	tags, err := h.Searcher.Search(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": tags,
	})
}
