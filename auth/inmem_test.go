package auth

import (
	"testing"
)

func TestInMemTeamRepository(t *testing.T) {
	repo := NewInMemTeamRepository()
	TestTeamRepository(t, repo)
}
