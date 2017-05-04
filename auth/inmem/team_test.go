package inmem

import (
	"testing"

	"github.com/bobinette/papernet/auth/testutil"
)

func TestInMemTeamRepository(t *testing.T) {
	repo := NewInMemTeamRepository()
	testutil.TestTeamRepository(t, repo)
}
