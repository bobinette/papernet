package imports

import (
	"context"

	"github.com/bobinette/papernet/errors"
)

var (
	ErrNotFound = errors.New("paper not found")
)

type Paper struct {
	ID int `json:"id"`

	Source    string `json:"source"`
	Reference string `json:"reference"`

	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Tags       []string `json:"tags"`
	Authors    []string `json:"authors"`
	References []string `json:"references"`
}

type Repository interface {
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
