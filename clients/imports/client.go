package imports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bobinette/papernet/clients/internal"
	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/users"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Decoder interface {
	Decode(v interface{}) error
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

func (c *Client) Search(ctx context.Context, q string, limit, offset int, sources []string) Decoder {
	user, err := users.FromContext(ctx)
	if err != nil {
		return internal.NewErrorDecoder(err)
	}

	token, err := internal.UserToken(user.ID, c.client, c.baseURL)
	if err != nil {
		return internal.NewErrorDecoder(err)
	}

	urlStr := fmt.Sprintf("%s/imports/v2/search", c.baseURL)
	u, err := url.Parse(urlStr)
	if err != nil {
		return internal.NewErrorDecoder(err)
	}

	qs := u.Query()
	qs.Set("q", q)
	qs.Set("limit", fmt.Sprintf("%d", limit))
	qs.Set("offset", fmt.Sprintf("%d", offset))
	qs["sources"] = sources
	u.RawQuery = qs.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return internal.NewErrorDecoder(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.client.Do(req)
	if err != nil {
		return internal.NewErrorDecoder(err)
	}

	if res.StatusCode != 200 {
		var callErr struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(res.Body).Decode(&callErr)
		if err != nil {
			return internal.NewErrorDecoder(err)
		}

		return internal.NewErrorDecoder(errors.New(
			fmt.Sprintf("error in call: %v", callErr.Message),
			errors.WithCode(res.StatusCode),
		))
	}

	return internal.NewJSONDecoder(res.Body)
}
