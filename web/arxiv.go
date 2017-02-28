package web

import (
	"net/http"
	"time"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
)

type ArxivHandler struct {
	Store papernet.PaperStore
	Index papernet.PaperIndex
}

func (h *ArxivHandler) Routes() []papernet.Route {
	return []papernet.Route{
		papernet.Route{
			Route:         "/arxiv",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: true,
			HandlerFunc:   WrapRequest(h.Search),
		},
	}
}

func (h *ArxivHandler) Search(req *Request) (interface{}, error) {
	user, err := auth.UserFromContext(req.Context())
	if err != nil {
		return nil, err
	}

	spider := papernet.ArxivSpider{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	arxivSearch := papernet.ArxivSearch{}
	err = req.Query("q", &arxivSearch.Q)
	if err != nil {
		return nil, err
	}
	err = req.Query("offset", &arxivSearch.Start)
	if err != nil {
		return nil, err
	}
	err = req.Query("limit", &arxivSearch.MaxResults)
	if err != nil {
		return nil, err
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
		"data":       papers,
		"pagination": res.Pagination,
	}, nil
}
