package auth

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

func (s *TeamService) Get(id int) (Team, error) {
	return Team{}, nil
}

func (s *TeamService) GetForUser(userID int) ([]Team, error) {
	return nil, nil
}

func (s *TeamService) Insert(creatorID int, team Team) (Team, error) {
	return Team{}, nil
}

func (s *TeamService) Invite(inviterID, teamID, invitedID int) (Team, error) {
	return Team{}, nil
}

func (s *TeamService) Kick(kickerID, teamID, kickeeID int) (Team, error) {
	return Team{}, nil
}

func (s *TeamService) Delete(deleterID, teamID int) error {
	return nil
}
