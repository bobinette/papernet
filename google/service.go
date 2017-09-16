package google

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"

	"github.com/bobinette/papernet/errors"
)

const (
	userInfoURL   = "https://www.googleapis.com/oauth2/v3/userinfo"
	userInfoScope = "https://www.googleapis.com/auth/userinfo.email"

	errInsufficientPermissions = "insufficientPermissions"

	papernetFolderName = "Papernet"
)

type googleUser struct {
	GoogleID string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type Service struct {
	repository UserRepository
	userClient *UserClient

	driveServiceFactory DriveServiceFactory

	// Config
	clientID     string
	clientSecret string
	redirectURL  string

	stateMutex sync.Locker
	state      map[string]string
}

func NewService(repo UserRepository, configPath string, dsf DriveServiceFactory, userClient *UserClient) (*Service, error) {
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

		driveServiceFactory: dsf,

		// Config
		clientID:     creds.ClientID,
		clientSecret: creds.ClientSecret,
		redirectURL:  creds.RedirectURL,

		stateMutex: &sync.RWMutex{},
		state:      make(map[string]string),
	}, nil
}

func (s *Service) LoginURL() string {
	state := randToken(32)
	s.stateMutex.Lock()
	s.state[state] = ""
	s.stateMutex.Unlock()

	url := s.config(userInfoScope).AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println(url)
	return url
}

func (s *Service) Login(state, code string) (string, string, error) {
	fromURL, ok := s.checkState(state)
	if !ok {
		return "", "", errors.New("invalid state", errors.BadRequest())
	}

	s.stateMutex.Lock()
	delete(s.state, state)
	s.stateMutex.Unlock()

	tok, err := s.config().Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", "", err
	}

	gUser, err := s.retrieveGoogleUser(tok)
	if err != nil {
		return "", "", err
	}

	user, err := s.repository.GetByGoogleID(gUser.GoogleID)
	if err != nil {
		return "", "", err
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
		return "", "", err
	} else if authUser.ID == 0 {
		return "", "", errors.New("user got no id")
	}

	user.ID = authUser.ID
	err = s.repository.Upsert(user)
	if err != nil {
		return "", "", err
	}

	token, err := s.userClient.Token(authUser)
	if err != nil {
		return "", "", err
	}

	return token, fromURL, nil
}

func (s *Service) hasDrive(userID int) (bool, error) {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return false, err
	}

	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, user.Token)
	driveService, err := s.driveServiceFactory(client)
	if err != nil {
		return false, fmt.Errorf("unable to retrieve drive Client %v\n", err)
	}

	return driveService.UserHasAllowedDrive()
}

func (s *Service) requireDrive(fromURL string) string {
	state := s.generateState(fromURL)
	config := s.config(userInfoScope, drive.DriveFileScope)
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (s *Service) inspectDrive(userID int, q string) ([]DriveFile, string, error) {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return nil, "", err
	}

	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, user.Token)
	ds, err := s.driveServiceFactory(client)
	if err != nil {
		return nil, "", fmt.Errorf("unable to create drive Client %v\n", err)
	}

	folderID, err := s.getOrCreateFolder(ds)
	if err != nil {
		return nil, "", err
	}

	return ds.ListFiles(folderID, q)
}

func (s *Service) addFile(userID int, filename, filetype string, data []byte) (DriveFile, error) {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return DriveFile{}, err
	}

	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, user.Token)
	ds, err := s.driveServiceFactory(client)
	if err != nil {
		return DriveFile{}, fmt.Errorf("unable to create drive Client %v\n", err)
	}

	folderID, err := s.getOrCreateFolder(ds)
	if err != nil {
		return DriveFile{}, err
	}

	return ds.CreateFile(filename, filetype, folderID, data)
}

func (s *Service) getOrCreateFolder(ds DriveService) (string, error) {
	folderID, err := ds.GetFolderID(papernetFolderName)
	if err != nil {
		return "", err
	} else if folderID != "" {
		return folderID, err
	}

	folderID, err = ds.CreateFolder(papernetFolderName)
	if err != nil {
		return "", err
	}
	return folderID, err
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

func (s *Service) generateState(fromURL string) string {
	state := randToken(32)
	s.stateMutex.Lock()
	s.state[state] = fromURL
	s.stateMutex.Unlock()

	go func() {
		time.Sleep(15 * time.Minute)
		s.stateMutex.Lock()
		delete(s.state, state)
		s.stateMutex.Unlock()
	}()

	return state
}

func (s *Service) checkState(state string) (string, bool) {
	s.stateMutex.Lock()
	fromURL, ok := s.state[state]
	delete(s.state, state)
	s.stateMutex.Unlock()
	return fromURL, ok
}
