package oauth

import (
	"github.com/bobinette/papernet/auth"
)

type UserService interface {
	Upsert(auth.User) (auth.User, error)
	Token(int) (string, error)
}

type UserClient struct {
	// We can keep calls internal for now
	service UserService
}

func NewUserClient(service UserService) *UserClient {
	return &UserClient{service: service}
}

func (s *UserClient) Upsert(user User) (User, error) {
	authUser := auth.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	authUser, err := s.service.Upsert(authUser)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:    authUser.ID,
		Name:  authUser.Name,
		Email: authUser.Email,
	}, nil
}

func (s *UserClient) Token(user User) (string, error) {
	return s.service.Token(user.ID)
}
