package oauth

import (
	"github.com/bobinette/papernet/auth"
)

type User struct {
	ID    string `json:"sub"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

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

func (s *UserClient) Upsert(user User) (string, error) {
	authUser := auth.User{
		Name:     user.Name,
		Email:    user.Email,
		GoogleID: user.ID,
	}

	authUser, err := s.service.Upsert(authUser)
	if err != nil {
		return "", err
	}
	return s.service.Token(authUser.ID)
}
