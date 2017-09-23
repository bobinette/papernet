package bolt

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bobinette/papernet/google"
)

func TestRepository(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	filename := tmpFile.Name()
	defer os.Remove(filename)

	driver := Driver{}
	err = driver.Open(filename)
	require.NoError(t, err)
	defer driver.Close()

	repo := NewUserRepository(&driver)
	google.TestRepository(t, repo)
}
