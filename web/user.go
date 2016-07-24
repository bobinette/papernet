package web

import (
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet/user"
)

type UserHandler struct {
	service *user.Service
	f       Formatter
}

func NewUserHandler(store *bolt.DB) (*UserHandler, error) {
	s, err := user.NewService(store)
	if err != nil {
		return nil, err
	}
	return &UserHandler{
		service: s,
		f:       Formatter{},
	}, nil
}

func (h *UserHandler) Register(r *gin.Engine) {
	r.POST("/users", h.f.Wrap(h.Create))
	r.GET("/users/:name", h.f.Wrap(h.Get))
	r.PUT("/users/:name", h.f.Wrap(h.Update))
}

func (h *UserHandler) Get(c *gin.Context) (interface{}, int, error) {
	name := c.Param("name")
	user, err := h.service.Get(name)
	if err != nil {
		return nil, http.StatusNotFound, err
	}
	return user, http.StatusOK, nil
}

func (h *UserHandler) Create(c *gin.Context) (interface{}, int, error) {
	var body struct {
		Username string `json:"name"`
	}
	err := c.BindJSON(&body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	user, err := h.service.Create(body.Username)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return user, http.StatusOK, nil
}

func (h *UserHandler) Update(c *gin.Context) (interface{}, int, error) {
	name := c.Param("name")

	var u user.User
	err := c.BindJSON(&u)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if name != u.Name {
		return nil, http.StatusBadRequest, fmt.Errorf("names do not match: %s (url) and %s (data)", name, u.Name)
	}

	err = h.service.Update(&u)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return u, http.StatusOK, nil
}
