package user

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type Service struct {
	repo *Repository
}

func NewService(store *bolt.DB) (*Service, error) {
	repo, err := NewRepository(store)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo: repo,
	}, nil
}

func (s *Service) Get(name string) (*User, error) {
	user, err := s.repo.Get(name)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, fmt.Errorf("No user with name %s", name)
	}

	return user, nil
}

func (s *Service) Create(name string) (*User, error) {
	user, err := s.repo.Get(name)
	if err != nil {
		return nil, err
	} else if user != nil {
		return nil, fmt.Errorf("user %s already exists", name)
	}

	user = &User{Name: name}
	err = s.repo.Upsert(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Update(user *User) error {
	err := s.repo.Upsert(user)
	if err != nil {
		return err
	}
	return nil
}
