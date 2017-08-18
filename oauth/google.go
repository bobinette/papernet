package oauth

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`

	Salt         string `json:"salt"`
	PasswordHash string `json:"password"`
}

type GoogleRepository interface {
	Get(googleID string) (int, error)
	Insert(googleID string, userID int) error
}
