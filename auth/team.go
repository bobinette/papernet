package auth

import (
	"fmt"

	"github.com/bobinette/papernet/errors"
)

func errTeamNotFound(id int) error {
	return errors.New(fmt.Sprintf("No team for id %d", id), errors.NotFound())
}

func errNotTeamAdmin(id int) error {
	return errors.New(fmt.Sprintf("You are not an admin of team %d", id), errors.Forbidden())
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

type TeamRepository interface {
	Get(int) (Team, error)
	GetForUser(int) ([]Team, error)

	Upsert(*Team) error
	Delete(int) error
}

type TeamService struct {
	repository TeamRepository
}

func NewTeamService(repo TeamRepository) *TeamService {
	return &TeamService{
		repository: repo,
	}
}

func (s *TeamService) Get(callerID int, teamID int) (Team, error) {
	team, err := s.repository.Get(teamID)
	if err != nil {
		return Team{}, err
	}

	// team.ID == 0 means that there was no team in the database
	if team.ID == 0 {
		return Team{}, errTeamNotFound(teamID)
	}

	// If the user is not a member of the team -> 404
	if !userIsMemberOfTeam(callerID, team) {
		return Team{}, errTeamNotFound(teamID)
	}

	return team, nil
}

func (s *TeamService) GetForUser(callerID int) ([]Team, error) {
	return s.repository.GetForUser(callerID)
}

func (s *TeamService) Insert(callerID int, team Team) (Team, error) {
	team.Members = []TeamMember{
		{ID: callerID, IsTeamAdmin: true},
	}
	team.CanSee = []int{}
	team.CanEdit = []int{}

	err := s.repository.Upsert(&team)
	if err != nil {
		return Team{}, err
	}

	return team, nil
}

func (s *TeamService) Invite(callerID, teamID, memberID int) (Team, error) {
	team, err := s.repository.Get(teamID)
	if err != nil {
		return Team{}, err
	}

	// team.ID == 0 means that there was no team in the database
	if team.ID == 0 {
		return Team{}, errTeamNotFound(teamID)
	}

	// If the user is not a member of the team -> 404
	if !userIsMemberOfTeam(callerID, team) {
		return Team{}, errTeamNotFound(teamID)
	}

	// If the user is not an admin of the team -> 403
	if !userIsAdminOfTeam(callerID, team) {
		return Team{}, errNotTeamAdmin(teamID)
	}

	team.Members = append(team.Members, TeamMember{ID: memberID, IsTeamAdmin: false})
	err = s.repository.Upsert(&team)
	if err != nil {
		return Team{}, err
	}

	return team, nil
}

func (s *TeamService) Kick(callerID, teamID, memberID int) (Team, error) {
	team, err := s.repository.Get(teamID)
	if err != nil {
		return Team{}, err
	}

	// team.ID == 0 means that there was no team in the database
	if team.ID == 0 {
		return Team{}, errTeamNotFound(teamID)
	}

	// If the user is not a member of the team -> 404
	if !userIsMemberOfTeam(callerID, team) {
		return Team{}, errTeamNotFound(teamID)
	}

	// If the user is not an admin of the team -> 403
	if callerID != memberID && !userIsAdminOfTeam(callerID, team) {
		return Team{}, errNotTeamAdmin(teamID)
	}

	index := -1
	for i, member := range team.Members {
		if member.ID == memberID {
			if member.IsTeamAdmin {
				return Team{}, errors.New("cannot kick team admin", errors.BadRequest())
			}
			index = i
			break
		}
	}

	if index == -1 {
		return Team{}, errors.New(fmt.Sprintf("user %d is not a member of team %d", memberID, teamID), errors.NotFound())
	} else if index == len(team.Members)-1 {
		team.Members = team.Members[0:index]
	} else {
		team.Members = append(team.Members[0:index], team.Members[index+1:len(team.Members)]...)
	}

	err = s.repository.Upsert(&team)
	if err != nil {
		return Team{}, err
	}

	return team, nil
}

func (s *TeamService) Delete(callerID, teamID int) error {
	team, err := s.repository.Get(teamID)
	if err != nil {
		return err
	}

	// If the user is not a member of the team -> 404
	if !userIsMemberOfTeam(callerID, team) {
		return errTeamNotFound(teamID)
	}

	// If the user is not an admin of the team -> 403
	if !userIsAdminOfTeam(callerID, team) {
		return errNotTeamAdmin(teamID)
	}

	return s.repository.Delete(team.ID)
}

func userIsMemberOfTeam(userID int, team Team) bool {
	for _, m := range team.Members {
		if m.ID == userID {
			return true
		}
	}
	return false
}

func userIsAdminOfTeam(userID int, team Team) bool {
	for _, m := range team.Members {
		if m.ID == userID {
			return m.IsTeamAdmin
		}
	}
	return false
}
