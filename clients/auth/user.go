package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	IsAdmin bool `json:"isAdmin"`

	Owns      []int `json:"owns"`
	CanSee    []int `json:"canSee"`
	CanEdit   []int `json:"canEdit"`
	Bookmarks []int `json:"bookmarks"`
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseURL string
	client  HTTPClient
}

func NewClient(c HTTPClient, baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  c,
	}
}

func (c *Client) User(id int) (User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/v2/users/%d", c.baseURL, id), nil)
	if err != nil {
		return User{}, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	var user User
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (c *Client) Token(id int) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/v2/users/%d/token", c.baseURL, id), nil)
	if err != nil {
		return "", err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var token struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
