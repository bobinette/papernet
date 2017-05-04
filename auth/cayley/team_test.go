package cayley

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bobinette/papernet/auth/testutil"
)

func createTeamRepository(t *testing.T) (*TeamRepository, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err, "could not create tmp file")

	filename := tmpFile.Name()
	store, err := NewStore(filename)
	require.NoError(t, err, "could not create store")

	repo := NewTeamRepository(store)
	return repo, func() {
		store.Close()
		os.Remove(filename)
	}
}

func TestTeamRepository(t *testing.T) {
	repo, tearDown := createTeamRepository(t)
	defer tearDown()

	testutil.TestTeamRepository(t, repo)
}
