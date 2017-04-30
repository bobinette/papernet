package auth

import (
	"sync"
)

type InMemTeamRepository struct {
	mu    sync.Locker
	teams []Team
	maxID int
}

func NewInMemTeamRepository() *InMemTeamRepository {
	return &InMemTeamRepository{
		mu:    &sync.Mutex{},
		teams: make([]Team, 0),
		maxID: 0,
	}
}

func (r *InMemTeamRepository) Get(id int) (Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, team := range r.teams {
		if team.ID == id {
			return team, nil
		}
	}
	return Team{}, nil
}

func (r *InMemTeamRepository) GetForUser(userID int) ([]Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	teams := make([]Team, 0)
	for _, team := range r.teams {
		for _, member := range team.Members {
			if member.ID == userID {
				teams = append(teams, team)
			}
		}
	}

	return teams, nil
}

func (r *InMemTeamRepository) Upsert(team *Team) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if team.ID == 0 {
		r.maxID++
		team.ID = r.maxID
	} else if team.ID > r.maxID {
		r.maxID = team.ID + 1
	}

	found := false
	for i, t := range r.teams {
		if t.ID == team.ID {
			r.teams[i] = *team
			found = true
			break
		}
	}
	if !found {
		r.teams = append(r.teams, *team)
	}

	return nil
}

func (r *InMemTeamRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	index := -1
	for i, team := range r.teams {
		if team.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return nil
	} else if index == len(r.teams)-1 {
		r.teams = r.teams[0:index]
	} else {
		r.teams = append(r.teams[0:index], r.teams[index+1:len(r.teams)]...)
	}

	return nil
}

type InMemUserRepository struct {
	mu    sync.Locker
	users []User
	maxID int

	teamRepository TeamRepository
}

func NewInMemUserRepository(teamRepo TeamRepository) *InMemUserRepository {
	return &InMemUserRepository{
		mu:    &sync.Mutex{},
		users: make([]User, 0),
		maxID: 0,

		teamRepository: teamRepo,
	}
}

func (r *InMemUserRepository) Get(userID int) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.get(userID)
}

func (r *InMemUserRepository) GetByGoogleID(googleID string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if user.GoogleID == googleID {
			return r.get(user.ID)
		}
	}

	return User{}, nil
}

func (r *InMemUserRepository) GetByEmail(email string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if user.Email == email {
			return r.get(user.ID)
		}
	}

	return User{}, nil
}

func (r *InMemUserRepository) List() ([]User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error
	users := make([]User, len(r.users))
	for i, user := range r.users {
		users[i], err = r.get(user.ID)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func (r *InMemUserRepository) get(userID int) (User, error) {
	var user User
	for _, u := range r.users {
		if u.ID == userID {
			user = u
		}
	}
	if user.ID == 0 {
		return User{}, nil
	}

	canSee := make(map[int]struct{})
	canEdit := make(map[int]struct{})
	for _, paperID := range user.Owns {
		canSee[paperID] = struct{}{}
		canEdit[paperID] = struct{}{}
	}

	teams, err := r.teamRepository.GetForUser(user.ID)
	if err != nil {
		return User{}, err
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

func (r *InMemUserRepository) Upsert(user *User) error {
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
