package services

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/errors"

	"github.com/bobinette/papernet/oauth"
)

var (
	googleEndpoint = google.Endpoint
	userInfoURL    = "https://www.googleapis.com/oauth2/v3/userinfo"
	scopes         = []string{
		"https://www.googleapis.com/auth/userinfo.email",
	}
)

type googleUser struct {
	GoogleID string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type GoogleService struct {
	repository oauth.GoogleRepository
	authClient *auth.Client
	config     oauth2.Config

	stateMutex sync.Locker
	state      map[string]struct{}
}

func NewGoogleService(repo oauth.GoogleRepository, configPath string, authClient *auth.Client) (*GoogleService, error) {
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
		repository: repo,
		authClient: authClient,
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
	state := randToken(32)
	s.stateMutex.Lock()
	s.state[state] = struct{}{}
	s.stateMutex.Unlock()

	return s.config.AuthCodeURL(state)
}

func (s *GoogleService) Login(state, code string) (string, error) {
	s.stateMutex.Lock()
	_, ok := s.state[state]
	s.stateMutex.Unlock() // no defer because the token exchange could be long

	if !ok {
		return "", errors.New("invalid state", errors.BadRequest())
	}

	s.stateMutex.Lock()
	delete(s.state, state)
	s.stateMutex.Unlock()

	tok, err := s.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", err
	}

	gUser, err := s.retrieveGoogleUser(tok)
	if err != nil {
		return "", err
	}

	userID, err := s.repository.Get(gUser.GoogleID)
	if err != nil {
		return "", err
	}

	user := oauth.User{
		ID:    userID,
		Name:  gUser.Name,
		Email: gUser.Email,
	}

	if user.ID == 0 {
		authUser := auth.User{
			Name:  user.Name,
			Email: user.Email,
		}
		authUser, err = s.authClient.Upsert(authUser)
		if err != nil {
			return "", err
		} else if authUser.ID == 0 {
			return "", errors.New("user got no id")
		}

		user.ID = authUser.ID
		err = s.repository.Insert(gUser.GoogleID, user.ID)
		if err != nil {
			return "", err
		}
	}

	return s.authClient.Token(user.ID)
}

func (s *GoogleService) retrieveGoogleUser(tok *oauth2.Token) (googleUser, error) {
	client := s.config.Client(oauth2.NoContext, tok)
	res, err := client.Get(userInfoURL)
	if err != nil {
		return googleUser{}, err
	}

	defer res.Body.Close()

	var user googleUser
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return googleUser{}, err
	}

	return user, nil
}
