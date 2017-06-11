package services

import (
	"context"
	"fmt"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/imports"
)

type PaperService interface {
	Insert(userID int, paper *imports.Paper, ctx context.Context) error
}

type ImportService struct {
	repository   imports.PaperRepository
	paperService PaperService
	searchers    []imports.Searcher
}

func NewImportService(repository imports.PaperRepository, paperService PaperService, searchers ...imports.Searcher) *ImportService {
	return &ImportService{
		repository:   repository,
		paperService: paperService,
		searchers:    searchers,
	}
}

func (s *ImportService) Sources() []string {
	sources := make([]string, len(s.searchers))
	for i, searcher := range s.searchers {
		sources[i] = searcher.Source()
	}

	return sources
}

func (s *ImportService) Import(userID int, paper imports.Paper, ctx context.Context) (imports.Paper, error) {
	err := s.paperService.Insert(userID, &paper, ctx)
	if err != nil {
		return imports.Paper{}, err
	} else if paper.ID == 0 {
		return imports.Paper{}, errors.New("id was not set when importing")
	}

	err = s.repository.Save(userID, paper.ID, paper.Source, paper.Reference)
	if err != nil {
		return imports.Paper{}, err
	}

	return paper, nil
}

func (s *ImportService) Search(
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

		// Get the ids of each entry from the repository
		for i, paper := range r.Papers {
			id, err := s.repository.Get(userID, searcher.Source(), paper.Reference)
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
