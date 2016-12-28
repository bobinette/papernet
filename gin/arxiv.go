package gin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

type ArxivHandler struct {
	Authenticator Authenticator

	Store papernet.PaperStore
	Index papernet.PaperIndex
}

func (h *ArxivHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/api/arxiv", JSONFormatter(h.Authenticator.Authenticate(h.Search)))
}

func (h *ArxivHandler) Search(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	spider := papernet.ArxivSpider{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	var start int
	startQ := c.Query("offset")
	if startQ != "" {
		start, err = strconv.Atoi(startQ)
		if err != nil {
			return nil, err
		}
	}

	var maxResults int
	maxResultsQ := c.Query("limit")
	if maxResultsQ != "" {
		maxResults, err = strconv.Atoi(maxResultsQ)
		if err != nil {
			return nil, err
		}
	}

	arxivSearch := papernet.ArxivSearch{
		Q:          c.Query("q"),
		Start:      start,
		MaxResults: maxResults,
	}

	res, err := spider.Search(arxivSearch)
	papers := res.Papers
	if err != nil {
		return nil, err
	}

	arxivIDs := make([]string, len(papers))
	for i, paper := range papers {
		arxivIDs[i] = paper.ArxivID
	}

	searchRes, err := h.Index.Search(papernet.PaperSearch{IDs: user.CanEdit, ArxivIDs: arxivIDs})
	if err != nil {
		return nil, err
	}

	savedPapers, err := h.Store.Get(searchRes.IDs...)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]int)
	for _, paper := range savedPapers {
		mapping[paper.ArxivID] = paper.ID
	}

	for _, paper := range papers {
		// If not in mapping (i.e. not imported yet), id is set to 0, so we are good
		paper.ID = mapping[paper.ArxivID]
	}

	return map[string]interface{}{
		"data":  papers,
		"total": res.Total,
	}, nil
}
