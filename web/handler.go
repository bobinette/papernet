package papernet

import (
	// "bytes"
	"fmt"
	// "io"
	// "io/ioutil"
	// "log"
	"net/http"
	// "os"
	// "os/exec"
	// "path"
	"strconv"
	// "strings"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet/database"
	// "github.com/bobinette/papernet/dot"
	"github.com/bobinette/papernet/models"
)

type WebHandler interface {
	Register(*gin.Engine)
}

type handler struct {
	db     database.DB
	search database.Search
}

func NewHandler(db database.DB, s database.Search) WebHandler {
	return &handler{
		db:     db,
		search: s,
	}
}

func (h *handler) Register(r *gin.Engine) {
	// ---- CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
		c.Next()
	})

	r.GET("/papers", h.list)
	r.POST("/papers", h.create)
	r.GET("/papers/:id", h.show)
	r.PUT("/papers/:id", h.update)
	r.DELETE("/papers/:id", h.delete)

	r.GET("/platform/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"data": "ok"})
	})
}

func (h *handler) show(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": fmt.Sprintf("invalid id: %s", c.Param("id")),
		})
		return
	}

	ps, err := h.db.Get(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": fmt.Sprintf("Error retrieving in db: %v", err),
		})
		return
	} else if len(ps) == 0 {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"message": fmt.Sprintf("No paper found for id %d", id),
		})
		return
	}
	p := ps[0]

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *handler) list(c *gin.Context) {
	q := c.Query("q")

	var ps []*models.Paper
	if q != "" {
		ids, err := h.search.Find(q)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"data":    "ko",
				"message": err.Error(),
			})
			return
		}
		ps, err = h.db.Get(ids...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"data":    "ko",
				"message": err.Error(),
			})
			return
		}
	} else {
		var err error
		ps, err = h.db.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"data":    "ko",
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": ps,
	})
}

func (h *handler) create(c *gin.Context) {
	var p models.Paper
	err := c.BindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"data":    "ko",
			"message": fmt.Sprintf("invalid data %v", err),
		})
		return
	}

	err = h.db.Insert(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"data":    "ko",
			"message": fmt.Sprintf("error inserting %v", err),
		})
		return
	}

	err = h.search.Index(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"data":    "ko",
			"message": fmt.Sprintf("error indexing %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *handler) update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"data": "ko", "message": "invalid id"})
		return
	}

	var p models.Paper
	err = c.BindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"data":    "ko",
			"message": fmt.Sprintf("invalid data %v", err),
		})
		return
	}

	if id != p.ID {
		c.JSON(http.StatusBadRequest, map[string]string{"data": "ko", "message": "ids do not match"})
		return
	}

	err = h.db.Update(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"data":    "ko",
			"message": fmt.Sprintf("error updating %v", err),
		})
		return
	}

	err = h.search.Index(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"data":    "ko",
			"message": fmt.Sprintf("error indexing %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *handler) delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"data": "ko", "message": "invalid id"})
		return
	}

	err = h.db.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"data":    "ko",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": "ok",
	})
}
