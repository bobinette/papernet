package paper

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type ListOptions struct {
	Search string
	IDs    []int
}

type Service struct {
	repo   *Repository
	search *Search
}

func NewService(store *bolt.DB, indexPath string) (*Service, error) {
	repo, err := NewRepository(store)
	if err != nil {
		return nil, err
	}

	search, err := NewSearch(indexPath)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo:   repo,
		search: search,
	}, nil
}

func (s *Service) Get(id int) (*Paper, error) {
	ps, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	} else if len(ps) == 0 {
		return nil, fmt.Errorf("no paper for id %d", id)
	}

	return ps[0], err
}

func (s *Service) List(opt ListOptions) ([]*Paper, error) {
	var ps []*Paper

	if opt.Search != "" {
		ids, err := s.search.Find(opt.Search)
		if err != nil {
			return nil, err
		}

		ps, err = s.repo.Get(ids...)
		if err != nil {
			return nil, err
		}
	} else if len(opt.IDs) > 0 {
		var err error
		ps, err = s.repo.Get(opt.IDs...)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		ps, err = s.repo.List()
		if err != nil {
			return nil, err
		}
	}

	return ps, nil
}

func (s *Service) Insert(p *Paper) error {
	err := s.repo.Insert(p)
	if err != nil {
		return err
	}

	err = s.search.Index(p)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Update(p *Paper) error {
	err := s.repo.Update(p)
	if err != nil {
		return err
	}

	err = s.search.Index(p)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}
