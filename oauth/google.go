package oauth

type GoogleRepository interface {
	Get(googleID string) (int, error)
	Insert(googleID string, userID int) error
}
