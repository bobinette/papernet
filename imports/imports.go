package imports

import (
	"context"
)

type Paper struct {
	ID int `json:"id"`

	Reference string   `json:"reference"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	Tags      []string `json:"tags"`
	Authors   []string `json:"authors"`
}

type PaperMapping interface {
	Save(userID, paperID int, source, ref string) error
	Get(userID int, source, ref string) (int, error)
}

type Pagination struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
	Total  uint `json:"total"`
}

type SearchResults struct {
	Papers     []Paper    `json:"papers"`
	Pagination Pagination `json:"pagination"`
}

type Searcher interface {
	Source() string
	Search(q string, limit, offset int, ctx context.Context) (SearchResults, error)
}
