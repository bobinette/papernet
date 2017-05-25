package oauth

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	Salt         string `json:"salt"`
	PasswordHash string `json:"password"`
}

type EmailRepository interface {
	Get(email string) (User, error)
	Insert(User) error
}
