package google

import (
	"github.com/bobinette/papernet/clients/auth"
)

type AuthUser struct {
	User

	Name  string `json:"-"`
	Email string `json:"-"`
}

type UserClient struct {
	// We can keep calls internal for now
	client *auth.Client
}

func NewUserClient(client *auth.Client) *UserClient {
	return &UserClient{
		client: client,
	}
}

func (c *UserClient) Upsert(user AuthUser) (AuthUser, error) {
	authUser := auth.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	authUser, err := c.client.Upsert(authUser)
	if err != nil {
		return AuthUser{}, err
	}
	return AuthUser{
		User: User{
			ID:       authUser.ID,
			GoogleID: user.GoogleID,
		},
		Name:  authUser.Name,
		Email: authUser.Email,
	}, nil
}

func (c *UserClient) Token(user AuthUser) (string, error) {
	return c.client.Token(user.ID)
}
