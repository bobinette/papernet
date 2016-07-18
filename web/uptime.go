package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UptimeHandler struct {
	f Formatter
}

func (h *UptimeHandler) Register(r *gin.Engine) {
	r.GET("/ping", h.f.Wrap(h.Ping))
}

func (h *UptimeHandler) Ping(c *gin.Context) (interface{}, int, error) {
	return "ok", http.StatusOK, nil
}
