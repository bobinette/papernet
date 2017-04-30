package auth

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T, repo UserRepository) {
	users := []*User{
		{
			Name:     "Pizza",
			Email:    "pizza@paper.net",
			GoogleID: "1234",
			IsAdmin:  true,
		},
		{
			Name:      "Yolo",
			Email:     "yolo@paper.net",
			GoogleID:  "5678",
			IsAdmin:   false,
			Owns:      []int{1, 10},
			CanSee:    []int{1, 10},
			CanEdit:   []int{1, 10},
			Bookmarks: []int{1, 10},
		},
	}

	// Insert user
	testInsertUser(t, repo, users)

	// Get users by id
	for i, user := range users {
		testGetUser(t, repo, user.ID, *user, fmt.Sprintf("get user %d", i))
	}

	// Get users by google id
	testGetUserByGoogleID(t, repo, users[1].GoogleID, *users[1], "get by google id")

	// Update pizza user email
	users[0].Email = "pizza@yolo.space"
	testUpdateUser(t, repo, users[0])

	// Update yolo user owns and bookmarks
	users[1].Owns = []int{1, 2, 3}
	users[1].Bookmarks = []int{1, 2}
	testUpdateUser(t, repo, users[1])

	// Test retrieving all the users

	// Retrieve paper owner
	testGetPaperOwner(t, repo, 1, users[1].ID)

	// Get by email
	testGetUserByEmail(t, repo, users[0].Email, *users[0], "get by email")

	// Delete user
	testDeleteUser(t, repo, users[1].ID, "delete")

	// Check there is no more owner for some paper
	testGetPaperOwner(t, repo, 1, 0)
}

func testInsertUser(t *testing.T, repo UserRepository, users []*User) {
	ids := make([]int, len(users))
	for i, user := range users {
		err := repo.Upsert(user)
		require.NoError(t, err, "insert %s must not fail", user.Name)
		require.NotEqual(t, 0, user.ID, "id must be set by insert")
		ids[i] = user.ID
	}

	// Test that all the ids are different
	sort.Ints(ids)
	for i := 0; i < len(ids)-1; i++ {
		require.NotEqual(t, ids[i], ids[i+1], "all ids must be different")
	}
}

func testGetUser(t *testing.T, repo UserRepository, id int, expected User, name string) {
	user, err := repo.Get(id)
	if assert.NoError(t, err, "%s - getting user should not fail", name) {
		assertUser(t, expected, user, name)
	}
}

func testGetUserByEmail(t *testing.T, repo UserRepository, email string, expected User, name string) {
	user, err := repo.GetByEmail(email)
	if assert.NoError(t, err, "%s - getting user by email should not fail", name) {
		assertUser(t, expected, user, name)
	}
}

func testGetUserByGoogleID(t *testing.T, repo UserRepository, googleID string, expected User, name string) {
	user, err := repo.GetByGoogleID(googleID)
	if assert.NoError(t, err, "%s - getting user by google id should not fail", name) {
		assertUser(t, expected, user, name)
	}
}

func testUpdateUser(t *testing.T, repo UserRepository, user *User) {
	id := user.ID
	err := repo.Upsert(user)
	assert.NoError(t, err, "%s - update should not have failed", user.Name)
	assert.Equal(t, id, user.ID, "id should not change")
}

func testDeleteUser(t *testing.T, repo UserRepository, userID int, name string) {
	err := repo.Delete(userID)
	assert.NoError(t, err, "%s - delete should not fail", name)

	retrieved, err := repo.Get(userID)
	assert.NoError(t, err, "%s - get after delete should not fail", name)
	assertUser(t, User{}, retrieved, name)
}

func testGetPaperOwner(t *testing.T, repo UserRepository, paperID, ownerID int) {
	userID, err := repo.PaperOwner(paperID)
	assert.NoError(t, err, "getting paper owner should not fail")
	assert.Equal(t, ownerID, userID, "incorrect owner id retrieved")
}

func testAllUsers(t *testing.T, repo UserRepository, users []*User) {
	retrieved, err := repo.List()
	if !assert.NoError(t, err, "listing all users should not fail") {
		return
	}

	if !assert.Equal(t, len(users), len(retrieved), "incorrect number of users retrieved") {
		return
	}

	for _, user := range users {
		found := false
		for _, retrievedUser := range retrieved {
			if retrievedUser.ID == user.ID {
				found = true
				assertUser(t, *user, retrievedUser, fmt.Sprintf("all - %s", user.Name))
			}
		}
		if !found {
			assert.Fail(t, "user %s not retrieved", user.Name)
		}
	}
}

func assertUser(t *testing.T, expected, actual User, name string) {
	// General information
	assert.Equal(t, expected.ID, actual.ID, "%s - ids should be equal", name)
	assert.Equal(t, expected.Name, actual.Name, "%s - names should be equal", name)
	assert.Equal(t, expected.Email, actual.Email, "%s - emails should be equal", name)
	assert.Equal(t, expected.GoogleID, actual.GoogleID, "%s - google ids should be equal", name)

	// Papers
	if assert.Equal(t, len(expected.Owns), len(actual.Owns), "%s - number of owned papers should be the same", name) {
		for _, paperID := range expected.Owns {
			assert.Contains(t, actual.Owns, paperID, "%s - paperID %d should be in owned papers", name, paperID)
		}
	}

	if assert.Equal(t, len(expected.CanSee), len(actual.CanSee), "%s - number of seeable papers should be the same", name) {
		for _, paperID := range expected.CanSee {
			assert.Contains(t, actual.CanSee, paperID, "%s - paperID %d should be in seeable papers", name, paperID)
		}
	}

	if assert.Equal(t, len(expected.CanEdit), len(actual.CanEdit), "%s - number of editable papers should be the same", name) {
		for _, paperID := range expected.CanEdit {
			assert.Contains(t, actual.CanEdit, paperID, "%s - paperID %d should be in editable papers", name, paperID)
		}
	}

	if assert.Equal(t, len(expected.Bookmarks), len(actual.Bookmarks), "%s - number of bookmarked papers should be the same", name) {
		for _, paperID := range expected.Bookmarks {
			assert.Contains(t, actual.Bookmarks, paperID, "%s - paperID %d should be in bookmarked papers", name, paperID)
		}
	}
}
