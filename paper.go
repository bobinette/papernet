package papernet

import (
	"time"
)

type Paper struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Authors []string `json:"authors"`

	Tags       []string `json:"tags"`
	References []string `json:"references"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// External ids
	ArxivID string `json:"arxivId"`
}

type Pagination struct {
	Total  uint64 `json:"total"`
	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

type PaperSearch struct {
	IDs []int  `json:"ids"`
	Q   string `json:"q"`

	Authors  []string `json:"authors"`
	Tags     []string `json:"tags"`
	ArxivIDs []string `json:"arxiv_ids"`

	Limit   uint64 `json:"limit"`
	Offset  uint64 `json:"offset"`
	OrderBy string `json:"orderBy"`
}

type PaperSearchTagsFacet []struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type PaperSearchFacets struct {
	Tags PaperSearchTagsFacet `json:"tags,omitempty"`
}

type PaperSearchResults struct {
	IDs        []int
	Facets     PaperSearchFacets
	Pagination Pagination
}

type PaperStore interface {
	Get(...int) ([]*Paper, error)
	List() ([]*Paper, error)
	Upsert(*Paper) error
	Delete(int) error
}

type PaperIndex interface {
	Index(*Paper) error
	Search(PaperSearch) (PaperSearchResults, error)
	Delete(int) error
}

type TagIndex interface {
	Index(string) error
	Search(string) ([]string, error)
}
