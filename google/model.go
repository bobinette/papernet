package google

import (
	"golang.org/x/oauth2"
)

type User struct {
	ID       int    `json:"id"`
	GoogleID string `json:"googleId"`

	Token *oauth2.Token `json:"token"`
}

type UserRepository interface {
	GetByID(id int) (User, error)
	GetByGoogleID(googleID string) (User, error)

	Upsert(User) error
}
