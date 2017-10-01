package imports

import (
	"context"
	"fmt"

	"github.com/bobinette/papernet/clients/paper"
	"github.com/bobinette/papernet/errors"
)

type Service struct {
	repository  Repository
	paperClient *paper.Client
	searchers   []Searcher
}

func NewService(repository Repository, paperClient *paper.Client, searchers ...Searcher) *Service {
	return &Service{
		repository:  repository,
		paperClient: paperClient,
		searchers:   searchers,
	}
}

func (s *Service) Sources() []string {
	sources := make([]string, len(s.searchers))
	for i, searcher := range s.searchers {
		sources[i] = searcher.Source()
	}

	return sources
}

func (s *Service) Import(ctx context.Context, userID int, p Paper) (Paper, error) {
	pp := paper.Paper{
		ID:      p.ID,
		Title:   p.Title,
		Summary: p.Summary,
		Tags:    p.Tags,

		Authors:    p.Authors,
		References: p.References,
	}

	var err error
	pp, err = s.paperClient.Insert(ctx, pp)
	if err != nil {
		return Paper{}, err
	} else if pp.ID == 0 {
		return Paper{}, errors.New("id was not set when importing")
	}

	p.ID = pp.ID

	err = s.repository.Save(userID, p.ID, p.Source, p.Reference)
	if err != nil {
		return Paper{}, err
	}

	return p, nil
}

func (s *Service) Search(
	ctx context.Context,
	userID int,
	q string,
	limit int,
	offset int,
	sources []string,
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
		r, err := searcher.Search(ctx, q, limit, offset)
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
