package services

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/bobinette/papernet/errors"

	"github.com/bobinette/papernet/oauth"
)

type EmailService struct {
	repository oauth.EmailRepository
	userClient *oauth.UserClient
}

func NewEmailService(repo oauth.EmailRepository, userClient *oauth.UserClient) *EmailService {
	return &EmailService{
		repository: repo,
		userClient: userClient,
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

	upsertedUser, err := s.userClient.Upsert(user)
	if err != nil {
		return "", err
	} else if upsertedUser.ID == 0 {
		return "", errors.New("user got no id")
	}

	user.ID = upsertedUser.ID
	err = s.repository.Insert(user)
	if err != nil {
		return "", err
	}

	return s.userClient.Token(user)
}

func (s *EmailService) Login(email, password string) (string, error) {
	user, err := s.repository.Get(email)
	fmt.Println(user.Salt)
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
	return s.userClient.Token(user)
}
