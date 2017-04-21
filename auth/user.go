package auth

import (
	"context"
	"fmt"

	"github.com/bobinette/papernet/errors"
)

var (
	errInvalidRequest = errors.New("invalid request")
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	GoogleID string `json:"-"`
}

type UserRepository interface {
	Get(int) (User, error)
	GetByGoogleID(string) (User, error)
	Upsert(*User) error
}

type Service struct {
	repository UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{
		repository: repo,
	}
}

type GetRequest struct {
	ID int
}

func (s *Service) Get(ctx context.Context, r interface{}) (interface{}, error) {
	req, ok := r.(GetRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	user, err := s.repository.Get(req.ID)
	if err != nil {
		return nil, err
	}

	if user.ID == 0 {
		return nil, errors.New(fmt.Sprintf("<User %d> not found", req.ID), errors.WithCode(404))
	}
	return user, nil
}

type UpsertRequest struct {
	Name     string
	Email    string
	GoogleID string
}

func (s *Service) Upsert(ctx context.Context, r interface{}) (interface{}, error) {
	req, ok := r.(UpsertRequest)
	if !ok {
		return nil, errInvalidRequest
	}

	user, err := s.repository.GetByGoogleID(req.GoogleID)
	if err != nil {
		return nil, err
	}

	// Update user details
	user.Name = req.Name
	user.Email = req.Email
	user.GoogleID = req.GoogleID

	err = s.repository.Upsert(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
