package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/bobinette/papernet/errors"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseURL string
	client  HTTPClient

	user     string
	password string

	mu    sync.Locker
	token string
}

func NewClient(user, password string, c HTTPClient, baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  c,

		user:     user,
		password: password,

		mu:    &sync.Mutex{},
		token: "",
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "" {
		token, err := c.getToken()
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// @TODO: handle expired token

	return c.client.Do(req)
}

func (c *Client) getToken() (string, error) {
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()

	if token == "" {
		err := c.authenticate()
		if err != nil {
			return "", err
		}

		c.mu.Lock()
		token = c.token
		c.mu.Unlock()
	}

	return token, nil
}

func (c *Client) refreshToken() (string, error) {
	c.mu.Lock()
	c.token = ""
	c.mu.Unlock()
	return c.getToken()
}

func (c *Client) authenticate() error {
	body := bytes.Buffer{}
	err := json.NewEncoder(&body).Encode(map[string]string{
		"email":    c.user,
		"password": c.password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/v2/login", c.baseURL), &body)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return errors.New(string(data))
	}

	var resBody struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(res.Body).Decode(&resBody)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.token = resBody.AccessToken
	c.mu.Unlock()
	return nil
}
