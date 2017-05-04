package inmem

import (
	"testing"

	"github.com/bobinette/papernet/auth/testutil"
)

func TestInMemUserRepository(t *testing.T) {
	teamRepo := NewInMemTeamRepository()
	repo := NewInMemUserRepository(teamRepo)
	testutil.TestUserRepository(t, repo)
}
