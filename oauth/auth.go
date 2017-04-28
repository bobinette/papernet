package oauth

import (
	"github.com/bobinette/papernet/auth"
)

type UserClient interface {
	// Upsert returns an interface because we do not care about the return
	// structure, it will just be forwarded on the login route
	Upsert(User) (interface{}, error)
}

type userClient struct {
	// We can keep calls internal for now
	service *auth.UserService
}

func NewUserClient(service *auth.UserService) UserClient {
	return &userClient{service: service}
}

func (s *userClient) Upsert(user User) (interface{}, error) {
	authUser := auth.User{
		Name:     user.Name,
		Email:    user.Email,
		GoogleID: user.ID,
	}

	return s.service.Upsert(authUser)
}
