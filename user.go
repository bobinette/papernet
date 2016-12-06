package papernet

type SigningKey struct {
	Key string `json:"k"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Bookmarks []int `json:"bookmarks"`

	CanSee  []int `json::"-"`
	CanEdit []int `json::"-"`
}

type UserRepository interface {
	Get(string) (*User, error)
	Upsert(*User) error
}
