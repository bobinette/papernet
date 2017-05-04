package auth

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
