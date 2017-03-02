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

type UserRepository interface {
	Get(string) (*User, error)
	Upsert(*User) error
}
