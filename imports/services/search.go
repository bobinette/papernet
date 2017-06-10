package services

import (
	"context"
	"fmt"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/imports"
)

type SearchService struct {
	mapping   imports.PaperMapping
	searchers []imports.Searcher
}

func NewSearchService(mapping imports.PaperMapping, searchers ...imports.Searcher) *SearchService {
	return &SearchService{
		mapping:   mapping,
		searchers: searchers,
	}
}

func (s *SearchService) Search(
	userID int,
	q string,
	limit int,
	offset int,
	sources []string,
	ctx context.Context,
) (map[string]imports.SearchResults, error) {
	// Select the searchers
	var searchers []imports.Searcher
	if len(sources) != 0 {
		searchers = make([]imports.Searcher, len(sources))
		for i, source := range sources {
			found := false
			for _, searcher := range s.searchers {
				if searcher.Source() == source {
					searchers[i] = searcher
					found = true
					break
				}
			}

			if !found {
				return nil, errors.New(fmt.Sprintf("unknown source %s", source), errors.BadRequest())
			}
		}
	} else {
		searchers = s.searchers
	}

	res := make(map[string]imports.SearchResults)
	for _, searcher := range searchers {
		// Use the searcher to fetch the data
		r, err := searcher.Search(q, limit, offset, ctx)
		if err != nil {
			return nil, err
		}

		// Get the ids of each entry from the mapping
		for i, paper := range r.Papers {
			id, err := s.mapping.Get(userID, searcher.Source(), paper.Reference)
			if err != nil {
				return nil, err
			}

			paper.ID = id
			r.Papers[i] = paper
		}

		res[searcher.Source()] = r
	}

	return res, nil
}

func isIn(str string, a []string) bool {
	for _, s := range a {
		if s == str {
			return true
		}
	}
	return false
}
