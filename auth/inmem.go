package auth

import (
	"sync"
)

type InMemTeamRepository struct {
	mu    sync.Locker
	Teams []Team
	maxID int
}

func NewInMemTeamRepository() *InMemTeamRepository {
	return &InMemTeamRepository{
		mu:    &sync.Mutex{},
		Teams: make([]Team, 0),
		maxID: 0,
	}
}

func (r *InMemTeamRepository) Get(id int) (Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, team := range r.Teams {
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
	for _, team := range r.Teams {
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
	}

	found := false
	for i, t := range r.Teams {
		if t.ID == team.ID {
			r.Teams[i] = *team
			found = true
			break
		}
	}
	if !found {
		r.Teams = append(r.Teams, *team)
	}

	return nil
}

func (r *InMemTeamRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	index := -1
	for i, team := range r.Teams {
		if team.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return nil
	} else if index == len(r.Teams)-1 {
		r.Teams = r.Teams[0:index]
	} else {
		r.Teams = append(r.Teams[0:index], r.Teams[index+1:len(r.Teams)]...)
	}

	return nil
}
