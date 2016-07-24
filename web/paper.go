package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet/paper"
)

type PaperHandler struct {
	service *paper.Service
	f       Formatter
}

func NewPaperHandler(store *bolt.DB, indexPath string) (*PaperHandler, error) {
	s, err := paper.NewService(store, indexPath)
	if err != nil {
		return nil, err
	}
	return &PaperHandler{
		service: s,
		f:       Formatter{},
	}, nil
}

func (h *PaperHandler) Register(r *gin.Engine) {
	r.GET("/papers/:id", h.f.Wrap(h.Get))
	r.GET("/papers", h.f.Wrap(h.List))
	r.POST("/papers", h.f.Wrap(h.Insert))
	r.PUT("/papers/:id", h.f.Wrap(h.Update))
	r.DELETE("/papers/:id", h.f.Wrap(h.Delete))
}

func (h *PaperHandler) Get(c *gin.Context) (interface{}, int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	p, err := h.service.Get(id)
	if err != nil {
		// Assume a not found. TODO: Carry code in error
		return nil, http.StatusNotFound, err
	}
	return p, http.StatusOK, nil
}

func (h *PaperHandler) List(c *gin.Context) (interface{}, int, error) {
	opt := paper.ListOptions{
		Search: c.Query("q"),
	}

	ps, err := h.service.List(opt)
	if err != nil {
		// Assume a not found. TODO: Carry code in error
		return nil, http.StatusNotFound, err
	}

	return ps, http.StatusOK, nil
}

func (h *PaperHandler) Insert(c *gin.Context) (interface{}, int, error) {
	var p paper.Paper
	err := c.BindJSON(&p)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	err = h.service.Insert(&p)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return p, http.StatusOK, nil
}

func (h *PaperHandler) Update(c *gin.Context) (interface{}, int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	var p paper.Paper
	err = c.BindJSON(&p)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if id != p.ID {
		return nil, http.StatusBadRequest, fmt.Errorf("ids do not match: %d (url) and %d (data)", id, p.ID)
	}

	err = h.service.Update(&p)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return p, http.StatusOK, nil
}

func (h *PaperHandler) Delete(c *gin.Context) (interface{}, int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	err = h.service.Delete(id)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return "ok", http.StatusOK, nil
}
