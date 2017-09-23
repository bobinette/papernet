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

type originInfo struct {
	fromURL string
	scopes  []string

	// @TODO: handle state expiration
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
	state      map[string]originInfo
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
		state:      make(map[string]originInfo),
	}, nil
}

func (s *Service) LoginURL() string {
	state := s.generateState("", userInfoScope)
	return s.config(userInfoScope).AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *Service) Login(state, code string) (string, string, error) {
	info, ok := s.checkState(state)
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

	if user.Tokens == nil {
		user.Tokens = make(map[string]*oauth2.Token)
	}

	for _, scope := range info.scopes {
		user.Tokens[scope] = tok
	}

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

	return token, info.fromURL, nil
}

func (s *Service) hasDrive(userID int) (bool, error) {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return false, err
	}

	token, ok := user.Tokens[drive.DriveFileScope]
	if !ok {
		return false, nil
	}

	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, token)
	driveService, err := s.driveServiceFactory(client)
	if err != nil {
		return false, fmt.Errorf("unable to retrieve drive Client %v\n", err)
	}

	return driveService.UserHasAllowedDrive()
}

func (s *Service) requireDrive(fromURL string) string {
	state := s.generateState(fromURL, drive.DriveFileScope, userInfoScope)
	config := s.config(drive.DriveFileScope, userInfoScope)
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *Service) inspectDrive(userID int, q string) ([]DriveFile, string, error) {
	user, err := s.repository.GetByID(userID)
	if err != nil {
		return nil, "", err
	}

	token := user.Tokens[drive.DriveFileScope]
	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, token)
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

	token := user.Tokens[drive.DriveFileScope]
	client := s.config(drive.DriveFileScope).Client(oauth2.NoContext, token)
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

func (s *Service) generateState(fromURL string, scopes ...string) string {
	state := randToken(32)
	s.stateMutex.Lock()
	s.state[state] = originInfo{
		fromURL: fromURL,
		scopes:  scopes,
	}
	s.stateMutex.Unlock()

	go func() {
		time.Sleep(15 * time.Minute)
		s.stateMutex.Lock()
		delete(s.state, state)
		s.stateMutex.Unlock()
	}()

	return state
}

func (s *Service) checkState(state string) (originInfo, bool) {
	s.stateMutex.Lock()
	info, ok := s.state[state]
	delete(s.state, state)
	s.stateMutex.Unlock()
	return info, ok
}
