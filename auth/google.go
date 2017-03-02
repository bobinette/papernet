package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/bobinette/papernet"
)

type GoogleClient struct {
	config oauth2.Config

	stateMutex sync.Locker
	state      map[string]struct{}
}

func NewGoogleClient(config string) (*GoogleClient, error) {
	c, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}

	creds := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURL  string `json:"redirect_url"`
	}{}
	err = json.Unmarshal(c, &creds)
	if err != nil {
		return nil, err
	}

	return &GoogleClient{
		config: oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			RedirectURL:  creds.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		},

		stateMutex: &sync.RWMutex{},
		state:      make(map[string]struct{}),
	}, nil
}

func (c *GoogleClient) LoginURL() string {
	s := randToken()
	c.stateMutex.Lock()
	c.state[s] = struct{}{}
	c.stateMutex.Unlock()

	url := c.config.AuthCodeURL(s)
	return url
}

func (c *GoogleClient) ExchangeToken(state, code string) (*papernet.User, error) {
	c.stateMutex.Lock()
	_, ok := c.state[state]
	c.stateMutex.Unlock()

	if !ok {
		return nil, errors.New("Invalid state")
	}

	c.stateMutex.Lock()
	delete(c.state, state)
	c.stateMutex.Unlock()

	tok, err := c.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	return c.userInfo(tok)
}

func (c *GoogleClient) UserInfo(token string) (*papernet.User, error) {
	tok := oauth2.Token{
		AccessToken: token,
	}
	return c.userInfo(&tok)
}

func (c *GoogleClient) userInfo(tok *oauth2.Token) (*papernet.User, error) {
	client := c.config.Client(oauth2.NoContext, tok)
	res, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var user struct {
		Sub   string `json:"sub"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	err = decoder.Decode(&user)
	if err != nil {
		return nil, err
	}

	return &papernet.User{
		ID:    user.Sub,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
