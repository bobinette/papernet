package inmem

type GoogleRepository struct {
	users map[string]int
}

func NewGoogleRepository() *GoogleRepository {
	return &GoogleRepository{
		users: make(map[string]int),
	}
}

func (r *GoogleRepository) Get(googleID string) (int, error) {
	return r.users[googleID], nil
}

func (r *GoogleRepository) Insert(googleID string, userID int) error {
	r.users[googleID] = userID
	return nil
}
