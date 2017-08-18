package paper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/users"
)

type Paper struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Authors []string `json:"authors"`

	Tags       []string `json:"tags"`
	References []string `json:"references"`
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

func (c *Client) Insert(ctx context.Context, p Paper) (Paper, error) {
	user, err := users.FromContext(ctx)
	if err != nil {
		return Paper{}, err
	}

	token, err := c.userToken(user.ID)
	if err != nil {
		return Paper{}, err
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(p); err != nil {
		return Paper{}, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/paper/v2/papers", c.baseURL), body)
	if err != nil {
		return Paper{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.client.Do(req)
	if err != nil {
		return Paper{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var callErr struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(res.Body).Decode(&callErr)
		if err != nil {
			return Paper{}, err
		}

		return Paper{}, errors.New(fmt.Sprintf("error in call: %v", err), errors.WithCode(res.StatusCode))
	}

	var rp struct {
		Data Paper `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&rp); err != nil {
		return Paper{}, err
	}

	return rp.Data, nil
}

func (c *Client) userToken(id int) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/v2/users/%d/token", c.baseURL, id), nil)
	if err != nil {
		return "", err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var callErr struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(res.Body).Decode(&callErr)
		if err != nil {
			return "", err
		}

		return "", errors.New(fmt.Sprintf("error in call: %v", err), errors.WithCode(res.StatusCode))
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
