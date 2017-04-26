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

	Owns      []int `json:"owns"`
	CanSee    []int `json:"canSee"`
	CanEdit   []int `json:"canEdit"`
	Bookmarks []int `json:"bookmarks"`
}

type TeamMember struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	IsTeamAdmin bool `json:"isTeamAdmin"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	Members []TeamMember `json:"members"`

	CanSee  []int `json:"canSee"`
	CanEdit []int `json:"canEdit"`
}

type UserRepository interface {
	// User information
	Get(int) (User, error)
	GetByGoogleID(string) (User, error)
	GetByEmail(string) (User, error)
	Upsert(*User) error

	// Paper ownership
	PaperOwner(paperID int) (int, error)
	UpdatePaperOwner(userID, paperID int, owns bool) error

	// Team membership
	GetTeam(int) (Team, error)
	UpsertTeam(*Team) error
	UserTeams(userID int) ([]Team, error)
	UpdateTeamMember(userID, teamID int, isMember, isAdmin bool) error

	// Team permissions
	UpdateTeamPermission(teamID, paperID int, canSee, canEdit bool) error

	// All the users
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
		return User{}, errors.New(
			fmt.Sprintf("<Paper %d> already has an owner", paperID),
			errors.WithCode(http.StatusForbidden),
		)
	}

	err = s.repository.UpdatePaperOwner(userID, paperID, owns)
	if err != nil {
		return User{}, err
	}

	// Get again to have updated user
	return s.repository.Get(userID)
}

// -----------------------------------------------------------------------------
// Teams

func (s *UserService) UserTeams(userID int) ([]Team, error) {
	return s.repository.UserTeams(userID)
}

func (s *UserService) InsertTeam(userID int, team Team) (Team, error) {
	if team.ID != 0 {
		return Team{}, errors.New("cannot update team via this handler", errors.WithCode(http.StatusBadRequest))
	}

	// No need for name and email
	team.Members = []TeamMember{
		TeamMember{
			ID:          userID,
			IsTeamAdmin: true,
		},
	}

	err := s.repository.UpsertTeam(&team)
	if err != nil {
		return Team{}, err
	}

	// Reload team to get everything
	return s.repository.GetTeam(team.ID)
}

func (s *UserService) UpdateTeamMember(inviterID int, memberEmail string, teamID int, isMember, isAdmin bool) (Team, error) {
	// Check that caller is admin of the team
	if inviterIsAdmin, err := s.userIsAdminOfTeam(inviterID, teamID); err != nil {
		return Team{}, err
	} else if !inviterIsAdmin {
		return Team{}, errors.New("Inviter should be admin of the team", errors.WithCode(http.StatusForbidden))
	}

	member, err := s.repository.GetByEmail(memberEmail)
	if err != nil {
		return Team{}, err
	} else if member.ID == 0 {
		return Team{}, errors.New(fmt.Sprintf("No user with email %s", memberEmail), errors.WithCode(http.StatusNotFound))
	}

	err = s.repository.UpdateTeamMember(member.ID, teamID, isMember, isAdmin)
	if err != nil {
		return Team{}, err
	}

	return s.repository.GetTeam(teamID)
}

func (s *UserService) SharePaper(userID, teamID, paperID int, canSee, canEdit bool) (Team, error) {
	err := s.repository.UpdateTeamPermission(teamID, paperID, canSee, canEdit)
	if err != nil {
		return Team{}, err
	}
	return s.repository.GetTeam(teamID)
}

func (s *UserService) userIsAdminOfTeam(userID, teamID int) (bool, error) {
	team, err := s.repository.GetTeam(teamID)
	if err != nil {
		return false, err
	} else if team.ID == 0 {
		return false, errors.New(fmt.Sprintf("<Team %d> not found", teamID), errors.WithCode(http.StatusNotFound))
	}

	for _, m := range team.Members {
		if m.ID == userID {
			return m.IsTeamAdmin, nil
		}
	}

	return false, nil
}

// -----------------------------------------------------------------------------
// List all users

func (s *UserService) List() ([]User, error) {
	return s.repository.List()
}
