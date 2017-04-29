package auth

import (
	"testing"
)

func TestInMemUserRepository(t *testing.T) {
	teamRepo := NewInMemTeamRepository()
	repo := NewInMemUserRepository(teamRepo)
	TestUserRepository(t, repo)
}

func TestInMemTeamRepository(t *testing.T) {
	repo := NewInMemTeamRepository()
	TestTeamRepository(t, repo)
}
