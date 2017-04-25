package auth

import (
	"fmt"
	"net/http"

	"github.com/bobinette/papernet/errors"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	GoogleID string `json:"googleID"`

	IsAdmin bool `json:"isAdmin"`

	Owns []int `json:"owns"`
}

type UserRepository interface {
	Get(int) (User, error)
	GetByGoogleID(string) (User, error)
	Upsert(*User) error

	PaperOwner(paperID int) (int, error)
	UpdatePaperOwner(userID, paperID int, owns bool) error

	List() ([]User, error)
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repository: repo,
	}
}

func (s *UserService) Get(id int) (User, error) {
	user, err := s.repository.Get(id)
	if err != nil {
		return User{}, err
	}

	if user.ID == 0 {
		return User{}, errors.New(fmt.Sprintf("<User %d> not found", id), errors.WithCode(http.StatusNotFound))
	}
	return user, nil
}

func (s *UserService) Upsert(u User) (User, error) {
	var user User
	if u.ID != 0 {
		var err error
		user, err = s.repository.Get(u.ID)
		if err != nil {
			return User{}, err
		} else if user.ID == 0 {
			return User{}, errors.New(fmt.Sprintf("<User %d> not found", u.ID), errors.WithCode(http.StatusNotFound))
		}
	} else {
		var err error
		user, err = s.repository.GetByGoogleID(u.GoogleID)
		if err != nil {
			return User{}, err
		}
	}

	// Update user details
	user.Name = u.Name
	user.Email = u.Email
	user.GoogleID = u.GoogleID

	// Because admin is always false from web, and we do not want to remove the privilege
	// every time an admin logs in
	// @TODO: find a way to remove admin privilege from a user.
	user.IsAdmin = user.IsAdmin || u.IsAdmin

	err := s.repository.Upsert(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *UserService) UpdateUserPapers(userID, paperID int, owns bool) (User, error) {
	user, err := s.repository.Get(userID)
	if err != nil {
		return User{}, err
	} else if user.ID == 0 {
		return User{}, errors.New(fmt.Sprintf("<User %d> not found", userID), errors.WithCode(http.StatusNotFound))
	}

	// @TODO: add a "transfer" parameter to transfer ownership if needed
	owner, err := s.repository.PaperOwner(paperID)
	if err != nil {
		return User{}, err
	}

	if owner != 0 && owner != userID {
		return User{}, errors.New(fmt.Sprintf("<Paper %d> already has an owner", paperID), errors.WithCode(http.StatusForbidden))
	}

	err = s.repository.UpdatePaperOwner(userID, paperID, owns)
	if err != nil {
		return User{}, err
	}

	// Get again to have updated user
	return s.repository.Get(userID)
}

func (s *UserService) List() ([]User, error) {
	return s.repository.List()
}
