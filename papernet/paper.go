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
}

type Pagination struct {
	Total  uint64 `json:"total"`
	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

type SearchParams struct {
	IDs  []int    `json:"ids"`
	Q    string   `json:"q"`
	Tags []string `json:"tags"`

	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

type TagsFacet []struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type Facets struct {
	Tags TagsFacet `json:"tags,omitempty"`
}

type SearchResults struct {
	IDs        []int
	Facets     Facets
	Pagination Pagination
}

type PaperRepository interface {
	Get(...int) ([]Paper, error)
	List() ([]Paper, error)
	Upsert(*Paper) error
	Delete(int) error
}

type PaperIndex interface {
	Index(*Paper) error
	Search(SearchParams) (SearchResults, error)
	Delete(int) error
}

type TagIndex interface {
	Index(string) error
	Search(string) ([]string, error)
}
