package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T, repo UserRepository) {
	user := User{
		ID:       1,
		GoogleID: "123",
	}

	err := repo.Upsert(user)
	assert.NoError(t, err)

	u, err := repo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user, u)

	u, err = repo.GetByGoogleID(user.GoogleID)
	assert.NoError(t, err)
	assert.Equal(t, user, u)
}
