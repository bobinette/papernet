package web

import (
	"github.com/bobinette/papernet"
)

type ImportHandler struct {
	Importer papernet.Importer
}

func (h *ImportHandler) Routes() []papernet.EndPoint {
	return []papernet.EndPoint{
		papernet.EndPoint{
			URL:           "/papernet/imports",
			Method:        "GET",
			Renderer:      "JSON",
			Authenticated: false,
			HandlerFunc:   WrapRequest(h.Import),
		},
		papernet.EndPoint{
			URL:           "/papernet/imports/chrome-bookmarks",
			Method:        "POST",
			Renderer:      "JSON",
			Authenticated: false,
			HandlerFunc:   WrapRequest(h.ImportChromeBookmarks),
		},
	}
}

func (h *ImportHandler) Import(req *Request) (interface{}, error) {
	var addr string
	err := req.Query("url", &addr)
	if err != nil {
		return nil, err
	}

	paper, err := h.Importer.Import(addr)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": paper,
	}, nil
}

func (h *ImportHandler) ImportChromeBookmarks(req *Request) (interface{}, error) {
	defer req.Body.Close()
	papers, err := papernet.ImportChromeBookmarks(req.Body)
	if err != nil {
		return nil, err
	}
	return papers, nil
}
