package cayley

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bobinette/papernet/auth/testutil"
)

// createRepository returns a user repository and a tearDown function for cleaning
func createRepository(t *testing.T) (*UserRepository, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err, "could not create tmp file")

	filename := tmpFile.Name()
	store, err := NewStore(filename)
	require.NoError(t, err, "could not create store")

	repo := NewUserRepository(store)
	return repo, func() {
		store.Close()
		os.Remove(filename)
	}
}

func TestUserRepository(t *testing.T) {
	repo, tearDown := createRepository(t)
	defer tearDown()

	testutil.TestUserRepository(t, repo)
}
