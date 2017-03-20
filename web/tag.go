package web

import (
	"github.com/bobinette/papernet"
)

type TagHandler struct {
	Searcher papernet.TagIndex
}

func (h *TagHandler) Routes() []papernet.EndPoint {
	return []papernet.EndPoint{
		papernet.EndPoint{
			URL:           "/tags",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: false,
			HandlerFunc:   WrapRequest(h.List),
		},
	}
}

func (h *TagHandler) List(req *Request) (interface{}, error) {
	var q string
	err := req.Query("q", &q)
	if err != nil {
		return nil, err
	}

	tags, err := h.Searcher.Search(q)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": tags,
	}, nil
}
