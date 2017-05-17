package auth

import (
	"github.com/bobinette/papernet/auth/services"
)

type Client struct {
	service *services.UserService
}

func NewClient(s *services.UserService) *Client {
	return &Client{
		service: s,
	}
}

func (c *Client) CreatePaper(userID, paperID int) error {
	_, err := c.service.CreatePaper(userID, paperID)
	return err
}
