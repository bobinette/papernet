package auth

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	Salt         string `json:"-"`
	PasswordHash string `json:"-"`

	IsAdmin bool `json:"isAdmin"`

	Owns      []int `json:"owns"`
	CanSee    []int `json:"canSee"`
	CanEdit   []int `json:"canEdit"`
	Bookmarks []int `json:"bookmarks"`
}

type UserRepository interface {
	// User information
	Get(int) (User, error)
	GetByEmail(string) (User, error)
	List() ([]User, error)
	Upsert(*User) error
	Delete(int) error

	// User -> Paper
	PaperOwner(paperID int) (int, error)
}
