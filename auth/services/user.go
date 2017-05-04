package services

import (
	"fmt"
	"net/http"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/errors"
)

type Encoder interface {
	Encode(int) (string, error)
}

type UserService struct {
	repository auth.UserRepository

	encoder Encoder
}

func NewUserService(repo auth.UserRepository, encoder Encoder) *UserService {
	return &UserService{
		repository: repo,
		encoder:    encoder,
	}
}

func (s *UserService) Get(id int) (auth.User, error) {
	user, err := s.repository.Get(id)
	if err != nil {
		return auth.User{}, err
	}

	if user.ID == 0 {
		return auth.User{}, errUserNotFound(id)
	}
	return user, nil
}

func (s *UserService) Upsert(u auth.User) (auth.User, error) {
	var user auth.User
	if u.ID != 0 {
		var err error
		user, err = s.repository.Get(u.ID)
		if err != nil {
			return auth.User{}, err
		} else if user.ID == 0 {
			return auth.User{}, errUserNotFound(u.ID)
		}
	} else {
		var err error
		user, err = s.repository.GetByGoogleID(u.GoogleID)
		if err != nil {
			return auth.User{}, err
		}
	}

	// Update user details
	user.Name = u.Name
	user.Email = u.Email
	user.GoogleID = u.GoogleID

	// Because admin is always false from web, and we do not want to remove the privilege
	// every time an admin logs in
	user.IsAdmin = user.IsAdmin || u.IsAdmin

	err := s.repository.Upsert(&user)
	if err != nil {
		return auth.User{}, err
	}

	return user, nil
}

func (s *UserService) CreatePaper(callerID, paperID int) (auth.User, error) {
	user, err := s.repository.Get(callerID)
	if err != nil {
		return auth.User{}, err
	} else if user.ID == 0 {
		return auth.User{}, errUserNotFound(callerID)
	}

	ownerID, err := s.repository.PaperOwner(paperID)
	if err != nil {
		return auth.User{}, err
	}

	if ownerID == callerID {
		return user, nil
	}
	if ownerID != 0 {
		return auth.User{}, errors.New(
			fmt.Sprintf("paper %d is already owned", paperID),
			errors.WithCode(http.StatusForbidden),
		)
	}

	user.Owns = append(user.Owns, paperID)
	err = s.repository.Upsert(&user)
	if err != nil {
		return auth.User{}, err
	}

	return user, nil
}

func (s *UserService) Bookmark(callerID, paperID int, bookmark bool) (auth.User, error) {
	user, err := s.repository.Get(callerID)
	if err != nil {
		return auth.User{}, err
	} else if user.ID == 0 {
		return auth.User{}, errUserNotFound(callerID)
	}

	// If the user cannot see the paper, consider it not found
	found := false
	for _, pID := range user.CanSee {
		if pID == paperID {
			found = true
			break
		}
	}
	if !found {
		return auth.User{}, errPaperNotFound(paperID)
	}

	index := -1
	for i, pID := range user.Bookmarks {
		if pID == paperID {
			index = i
			break
		}
	}

	if !bookmark { // Remove bookmark
		if index == -1 {
			return user, nil
		}

		if index == len(user.Bookmarks)-1 {
			user.Bookmarks = user.Bookmarks[0:index]
		} else {
			user.Bookmarks = append(user.Bookmarks[0:index], user.Bookmarks[index+1:len(user.Bookmarks)-1]...)
		}
	} else { // Add bookmark
		if index != -1 {
			return user, nil
		}
		user.Bookmarks = append(user.Bookmarks, paperID)
	}

	err = s.repository.Upsert(&user)
	if err != nil {
		return auth.User{}, err
	}

	return user, nil
}

func (s *UserService) Token(userID int) (string, error) {
	return s.encoder.Encode(userID)
}

func (s *UserService) All() ([]auth.User, error) {
	return s.repository.List()
}
