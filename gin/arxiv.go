package gin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

type ArxivHandler struct {
	Authenticator Authenticator
}

func (h *ArxivHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/api/arxiv", JSONFormatter(h.Authenticator.Authenticate(h.Search)))
}

func (h *ArxivHandler) Search(c *gin.Context) (interface{}, error) {
	spider := papernet.ArxivSpider{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	papers, err := spider.Search(c.Query("q"))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": papers,
	}, nil
}
