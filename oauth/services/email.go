package services

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/errors"

	"github.com/bobinette/papernet/oauth"
)

type EmailService struct {
	repository oauth.EmailRepository
	authClient *auth.Client
}

func NewEmailService(repo oauth.EmailRepository, authClient *auth.Client) *EmailService {
	return &EmailService{
		repository: repo,
		authClient: authClient,
	}
}

func (s *EmailService) SignUp(email, password string) (string, error) {
	user, err := s.repository.Get(email)
	if err != nil {
		return "", err
	} else if user.ID != 0 {
		return "", errors.New("email already exists", errors.BadRequest())
	}

	user = oauth.User{
		Name:  email,
		Email: email,
		Salt:  randToken(64),
	}

	// Generate "hash" to store from user password
	hash, err := bcrypt.GenerateFromPassword([]byte(password+user.Salt), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user.PasswordHash = string(hash)

	authUser := auth.User{
		Name:  email,
		Email: email,
	}
	authUser, err = s.authClient.Upsert(authUser)
	if err != nil {
		return "", err
	} else if authUser.ID == 0 {
		return "", errors.New("user got no id")
	}

	user.ID = authUser.ID
	err = s.repository.Insert(user)
	if err != nil {
		return "", err
	}

	return s.authClient.Token(user.ID)
}

func (s *EmailService) Login(email, password string) (string, error) {
	user, err := s.repository.Get(email)
	if err != nil {
		return "", err
	} else if user.ID == 0 {
		return "", errors.New("email or password incorrect", errors.BadRequest())
	}

	// Comparing the password with the hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+user.Salt)); err != nil {
		return "", errors.New("email or password incorrect", errors.BadRequest())
	}

	// Password is correct here
	return s.authClient.Token(user.ID)
}
