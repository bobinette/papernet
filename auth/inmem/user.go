package inmem

import (
	"sync"

	"github.com/bobinette/papernet/auth"
)

type InMemUserRepository struct {
	mu    sync.Locker
	users []auth.User
	maxID int

	teamRepository auth.TeamRepository
}

func NewInMemUserRepository(teamRepo auth.TeamRepository) *InMemUserRepository {
	return &InMemUserRepository{
		mu:    &sync.Mutex{},
		users: make([]auth.User, 0),
		maxID: 0,

		teamRepository: teamRepo,
	}
}

func (r *InMemUserRepository) Get(userID int) (auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.get(userID)
}

func (r *InMemUserRepository) GetByGoogleID(googleID string) (auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if user.GoogleID == googleID {
			return r.get(user.ID)
		}
	}

	return auth.User{}, nil
}

func (r *InMemUserRepository) GetByEmail(email string) (auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if user.Email == email {
			return r.get(user.ID)
		}
	}

	return auth.User{}, nil
}

func (r *InMemUserRepository) List() ([]auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error
	users := make([]auth.User, len(r.users))
	for i, user := range r.users {
		users[i], err = r.get(user.ID)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func (r *InMemUserRepository) get(userID int) (auth.User, error) {
	var user auth.User
	for _, u := range r.users {
		if u.ID == userID {
			user = u
		}
	}
	if user.ID == 0 {
		return auth.User{}, nil
	}

	canSee := make(map[int]struct{})
	canEdit := make(map[int]struct{})
	for _, paperID := range user.Owns {
		canSee[paperID] = struct{}{}
		canEdit[paperID] = struct{}{}
	}

	teams, err := r.teamRepository.GetForUser(user.ID)
	if err != nil {
		return auth.User{}, err
	}

	for _, team := range teams {
		for _, paperID := range team.CanSee {
			canSee[paperID] = struct{}{}
		}

		for _, paperID := range team.CanEdit {
			canEdit[paperID] = struct{}{}
		}
	}

	user.CanSee = make([]int, 0)
	user.CanEdit = make([]int, 0)
	for paperID := range canSee {
		user.CanSee = append(user.CanSee, paperID)
	}

	for paperID := range canEdit {
		user.CanEdit = append(user.CanEdit, paperID)
	}

	return user, nil
}

func (r *InMemUserRepository) Upsert(user *auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == 0 {
		r.maxID++
		user.ID = r.maxID
	} else if user.ID > r.maxID {
		r.maxID = user.ID + 1
	}

	found := false
	for i, u := range r.users {
		if user.ID == u.ID {
			r.users[i] = *user
			found = true
			break
		}
	}

	if !found {
		r.users = append(r.users, *user)
	}

	return nil
}

func (r *InMemUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	index := -1
	for i, user := range r.users {
		if user.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return nil
	} else if index == len(r.users)-1 {
		r.users = r.users[0:index]
	} else {
		r.users = append(r.users[0:index], r.users[index+1:len(r.users)]...)
	}

	return nil
}

func (r *InMemUserRepository) PaperOwner(paperID int) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		for _, ownedID := range user.Owns {
			if ownedID == paperID {
				return user.ID, nil
			}
		}
	}

	return 0, nil
}
