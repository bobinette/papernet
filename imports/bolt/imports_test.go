package bolt

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setUp(t *testing.T) (*Driver, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err, "tmp file")

	filename := tmpFile.Name()
	driver := &Driver{}
	err = driver.Open(filename)
	require.NoError(t, err, "open driver")

	return driver, func() {
		driver.Close()
		os.Remove(filename)
	}
}

func TestPaperRepository(t *testing.T) {
	driver, tearDown := setUp(t)
	defer tearDown()

	repo := NewPaperRepository(driver)

	// Not inserted yet -> id is 0
	id, err := repo.Get(1, "source 1", "ref 1")
	require.NoError(t, err, "get non inserted u1 s1 r1")
	assert.Equal(t, 0, id, "get non inserted u1 s1 r1 - id")

	err = repo.Save(1, 10, "source 1", "ref 1")
	require.NoError(t, err, "insert u1 p10 s1 r1")

	id, err = repo.Get(1, "source 1", "ref 1")
	require.NoError(t, err, "get u1 s1 r1")
	assert.Equal(t, 10, id, "get u1 s1 r1 - id")
}
