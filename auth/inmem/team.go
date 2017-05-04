package inmem

import (
	"sync"

	"github.com/bobinette/papernet/auth"
)

type InMemTeamRepository struct {
	mu    sync.Locker
	teams []auth.Team
	maxID int
}

func NewInMemTeamRepository() *InMemTeamRepository {
	return &InMemTeamRepository{
		mu:    &sync.Mutex{},
		teams: make([]auth.Team, 0),
		maxID: 0,
	}
}

func (r *InMemTeamRepository) Get(id int) (auth.Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, team := range r.teams {
		if team.ID == id {
			return team, nil
		}
	}
	return auth.Team{}, nil
}

func (r *InMemTeamRepository) GetForUser(userID int) ([]auth.Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	teams := make([]auth.Team, 0)
	for _, team := range r.teams {
		for _, member := range team.Members {
			if member.ID == userID {
				teams = append(teams, team)
			}
		}
	}

	return teams, nil
}

func (r *InMemTeamRepository) Upsert(team *auth.Team) error {
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
