package oauth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleEndpoint = google.Endpoint
	userInfoURL    = "https://www.googleapis.com/oauth2/v3/userinfo"
	scopes         = []string{
		"https://www.googleapis.com/auth/userinfo.email",
	}
)

type GoogleService struct {
	userClient UserClient
	config     oauth2.Config

	stateMutex sync.Locker
	state      map[string]struct{}
}

type User struct {
	ID    string `json:"sub"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewGoogleService(configPath string, userClient UserClient) (*GoogleService, error) {
	c, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var creds struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURL  string `json:"redirect_url"`
	}
	err = json.Unmarshal(c, &creds)
	if err != nil {
		return nil, err
	}

	return &GoogleService{
		userClient: userClient,
		config: oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			RedirectURL:  creds.RedirectURL,
			Scopes:       scopes,
			Endpoint:     googleEndpoint,
		},

		stateMutex: &sync.RWMutex{},
		state:      make(map[string]struct{}),
	}, nil
}

func (s *GoogleService) LoginURL() string {
	state := randToken()
	s.stateMutex.Lock()
	s.state[state] = struct{}{}
	s.stateMutex.Unlock()

	return s.config.AuthCodeURL(state)
}

func (s *GoogleService) Login(state, code string) (interface{}, error) {
	s.stateMutex.Lock()
	_, ok := s.state[state]
	s.stateMutex.Unlock() // no defer because the token exchange could be long

	if !ok {
		return nil, errors.New("Invalid state")
	}

	s.stateMutex.Lock()
	delete(s.state, state)
	s.stateMutex.Unlock()

	tok, err := s.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	user, err := s.retrieveUser(tok)
	if err != nil {
		return nil, err
	}

	return s.userClient.Upsert(user)
}

func (s *GoogleService) retrieveUser(tok *oauth2.Token) (User, error) {
	client := s.config.Client(oauth2.NoContext, tok)
	res, err := client.Get(userInfoURL)
	if err != nil {
		return User{}, err
	}

	defer res.Body.Close()

	var user User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
