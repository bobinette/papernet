package papernet

type SigningKey struct {
	Key string `json:"k"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	Bookmarks []int `json:"bookmarks"`

	CanSee  []int `json:"canSee"`
	CanEdit []int `json:"canEdit"`
}

type UserStore interface {
	Get(string) (*User, error)
	Upsert(*User) error
}

type PermissionManager interface {
	UserCanSee(string, int) (bool, error)
	UserCanEdit(string, int) (bool, error)

	AllowUserToSee(string, int) error
	AllowUserToEdit(string, int) error
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	Admins  []string `json:"admins"`
	Members []string `json:"members"`

	CanSee  []int `json:"canSee"`
	CanEdit []int `json:"canEdit"`
}

type TeamStore interface {
	Get(int) (Team, error)
	Upsert(*Team) error
	Delete(int) error

	List(userID string) ([]Team, error)
}
