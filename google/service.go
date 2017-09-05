package google

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"

	"github.com/bobinette/papernet/errors"
)

var (
	userInfoURL   = "https://www.googleapis.com/oauth2/v3/userinfo"
	userInfoScope = "https://www.googleapis.com/auth/userinfo.email"

	errInsufficientPermissions = "insufficientPermissions"
)

type googleUser struct {
	GoogleID string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type Service struct {
	repository UserRepository
	userClient *UserClient

	// Config
	clientID     string
	clientSecret string
	redirectURL  string

	stateMutex sync.Locker
	state      map[string]struct{}
}

func NewService(repo UserRepository, configPath string, userClient *UserClient) (*Service, error) {
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

	return &Service{
		repository: repo,
		userClient: userClient,

		// Config
		clientID:     creds.ClientID,
		clientSecret: creds.ClientSecret,
		redirectURL:  creds.RedirectURL,

		stateMutex: &sync.RWMutex{},
		state:      make(map[string]struct{}),
	}, nil
}

func (s *Service) LoginURL() string {
	state := randToken(32)
	s.stateMutex.Lock()
	s.state[state] = struct{}{}
	s.stateMutex.Unlock()

	return s.config(userInfoScope).AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (s *Service) Login(state, code string) (string, error) {
	s.stateMutex.Lock()
	_, ok := s.state[state]
	s.stateMutex.Unlock() // no defer because the token exchange could be long

	if !ok {
		return "", errors.New("invalid state", errors.BadRequest())
	}

	s.stateMutex.Lock()
	delete(s.state, state)
	s.stateMutex.Unlock()

	tok, err := s.config().Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", err
	}

	gUser, err := s.retrieveGoogleUser(tok)
	if err != nil {
		return "", err
	}

	user, err := s.repository.GetByGoogleID(gUser.GoogleID)
	if err != nil {
		return "", err
	} else if user.ID == 0 {
		user.GoogleID = gUser.GoogleID
	}

	user.Token = tok

	authUser := AuthUser{
		User:  user,
		Name:  gUser.Name,
		Email: gUser.Email,
	}

	// Update infos in auth
	authUser, err = s.userClient.Upsert(authUser)
	if err != nil {
		return "", err
	} else if authUser.ID == 0 {
		return "", errors.New("user got no id")
	}

	user.ID = authUser.ID
	err = s.repository.Upsert(user)
	if err != nil {
		return "", err
	}

	return s.userClient.Token(authUser)
}

func (s *Service) requireDrive(userID int) (bool, string, error) {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return false, "", err
	}

	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, user.Token)

	driveClient, err := drive.New(client)
	if err != nil {
		return false, "", fmt.Errorf("unable to retrieve drive Client %v\n", err)
	}

	_, err = driveClient.Files.List().PageSize(10).
		Fields("nextPageToken, files(id, name)").Do()

	// No error means the user already has access to the drive
	if err == nil {
		return true, "", nil
	}

	if err != nil {
		isInsufficientPermission := false
		code := 500
		if err, ok := err.(*googleapi.Error); ok && err.Code == 403 {
			code = err.Code
			for _, e := range err.Errors {
				if e.Reason == errInsufficientPermissions {
					isInsufficientPermission = true
					break
				}
			}
		}

		// The error is something else, returning to the caller
		if !isInsufficientPermission {
			return false, "", errors.New("unable to retrieve files: %v\n", errors.WithCause(err), errors.WithCode(code))
		}
	}

	// The user did not granted the app enough permissions, require them
	state := randToken(32)
	s.stateMutex.Lock()
	s.state[state] = struct{}{}
	s.stateMutex.Unlock()

	url := s.config(userInfoScope, drive.DriveFileScope).AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return false, url, nil
}

func (s *Service) inspectDrive(userID int) error {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return err
	}
	fmt.Printf("User: %+v\n", user)

	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, user.Token)

	driveClient, err := drive.New(client)
	if err != nil {
		return fmt.Errorf("unable to retrieve drive Client %v\n", err)
	}

	r, err := driveClient.Files.List().PageSize(10).
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		if err, ok := err.(*googleapi.Error); ok && err.Code == 403 {
			fmt.Printf("%+v\n", err.Body)
		}
		return fmt.Errorf("unable to retrieve files: %v\n", err)
	}

	fmt.Println("Files:")
	if len(r.Files) > 0 {
		for _, i := range r.Files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	} else {
		fmt.Println("No files found.")
	}
	return nil
}

func (s *Service) config(scopes ...string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		RedirectURL:  s.redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

func (s *Service) retrieveGoogleUser(tok *oauth2.Token) (googleUser, error) {
	client := s.config().Client(oauth2.NoContext, tok)
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
