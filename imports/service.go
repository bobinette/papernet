package imports

import (
	"context"
	"fmt"

	"github.com/bobinette/papernet/errors"
)

type PaperService interface {
	Insert(userID int, paper *Paper, ctx context.Context) error
}

type Service struct {
	repository   Repository
	paperService PaperService
	searchers    []Searcher
}

func NewService(repository Repository, paperService PaperService, searchers ...Searcher) *Service {
	return &Service{
		repository:   repository,
		paperService: paperService,
		searchers:    searchers,
	}
}

func (s *Service) Sources() []string {
	sources := make([]string, len(s.searchers))
	for i, searcher := range s.searchers {
		sources[i] = searcher.Source()
	}

	return sources
}

func (s *Service) Import(userID int, paper Paper, ctx context.Context) (Paper, error) {
	err := s.paperService.Insert(userID, &paper, ctx)
	if err != nil {
		return Paper{}, err
	} else if paper.ID == 0 {
		return Paper{}, errors.New("id was not set when importing")
	}

	err = s.repository.Save(userID, paper.ID, paper.Source, paper.Reference)
	if err != nil {
		return Paper{}, err
	}

	return paper, nil
}

func (s *Service) Search(
	userID int,
	q string,
	limit int,
	offset int,
	sources []string,
	ctx context.Context,
) (map[string]SearchResults, error) {
	// Select the searchers
	var searchers []Searcher
	if len(sources) != 0 {
		searchers = make([]Searcher, len(sources))
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

	res := make(map[string]SearchResults)
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
