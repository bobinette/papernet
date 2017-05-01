package auth

import (
	"fmt"

	"github.com/bobinette/papernet/errors"
)

type TeamMember struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	IsTeamAdmin bool `json:"admin"`
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
	repository     TeamRepository
	userRepository UserRepository
}

func NewTeamService(repo TeamRepository, userRepo UserRepository) *TeamService {
	return &TeamService{
		repository:     repo,
		userRepository: userRepo,
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

func (s *TeamService) Create(callerID int, team Team) (Team, error) {
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

func (s *TeamService) Invite(callerID, teamID int, memberEmail string) (Team, error) {
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

	user, err := s.userRepository.GetByEmail(memberEmail)
	if err != nil {
		return Team{}, err
	} else if user.ID == 0 {
		return Team{}, errors.New(fmt.Sprintf("no user found for email %s", memberEmail), errors.NotFound())
	}

	if userIsMemberOfTeam(user.ID, team) {
		return team, nil
	}

	team.Members = append(team.Members, TeamMember{ID: user.ID, IsTeamAdmin: false})
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

func (s *TeamService) Share(callerID, teamID, paperID int, canEdit bool) (Team, error) {
	user, err := s.userRepository.Get(callerID)
	if err != nil {
		return Team{}, err
	} else if user.ID == 0 {
		return Team{}, errUserNotFound(callerID)
	}

	found := false
	for _, pID := range user.CanSee {
		if pID == paperID {
			found = true
			break
		}
	}
	if !found {
		return Team{}, errPaperNotFound(paperID)
	}

	found = false
	for _, pID := range user.Owns {
		if pID == paperID {
			found = true
			break
		}
	}
	if !found {
		return Team{}, errors.New(fmt.Sprintf("you cannot share paper %d because you are not the owner", paperID), errors.Forbidden())
	}

	team, err := s.repository.Get(teamID)
	if err != nil {
		return Team{}, err
	} else if team.ID == 0 {
		return Team{}, errTeamNotFound(teamID)
	}

	// If the user is not a member of the team -> 404
	if !userIsMemberOfTeam(callerID, team) {
		return Team{}, errTeamNotFound(teamID)
	}

	found = false
	for _, canSeeID := range team.CanSee {
		if canSeeID == paperID {
			found = true
			fmt.Println(team.CanSee, paperID)
			break
		}
	}
	if !found {
		team.CanSee = append(team.CanSee, paperID)
	}

	if canEdit {
		found = false
		for _, canEditID := range team.CanEdit {
			if canEditID == paperID {
				found = true
				break
			}
		}
		if !found {
			team.CanEdit = append(team.CanEdit, paperID)
		}
	}

	err = s.repository.Upsert(&team)
	if err != nil {
		return Team{}, err
	}

	return team, nil
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
