package google

type InmemRepository struct {
	users []User
}

func NewInmemRepository() *InmemRepository {
	return &InmemRepository{
		users: make([]User, 0),
	}
}

func (r InmemRepository) GetByID(id int) (User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return User{}, nil
}

func (r InmemRepository) GetByGoogleID(googleID string) (User, error) {
	for _, user := range r.users {
		if user.GoogleID == googleID {
			return user, nil
		}
	}
	return User{}, nil
}

func (r *InmemRepository) Upsert(user User) error {
	for i, u := range r.users {
		if u.ID == user.ID {
			r.users[i] = user
			return nil
		}
	}

	r.users = append(r.users, user)
	return nil
}
