package oauth

import (
	"github.com/bobinette/papernet/auth"
)

type UserClient interface {
	// Upsert returns an interface because we do not care about the return
	// structure, it will just be forwarded on the login route
	Upsert(User) (string, error)
}

type userClient struct {
	// We can keep calls internal for now
	service UserService
}

type UserService interface {
	Upsert(auth.User) (auth.User, error)
	Token(int) (string, error)
}

func NewUserClient(service UserService) UserClient {
	return &userClient{service: service}
}

func (s *userClient) Upsert(user User) (string, error) {
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
