package auth

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRepository(t *testing.T, repo TeamRepository) {
	teams := []*Team{
		{
			Name: "Pizza",
			Members: []TeamMember{
				{ID: 1, IsTeamAdmin: true},
				{ID: 2, IsTeamAdmin: false},
			},
			CanSee:  []int{1, 2},
			CanEdit: []int{1},
		},
		{
			Name: "Yolo",
			Members: []TeamMember{
				{ID: 1, IsTeamAdmin: false},
				{ID: 3, IsTeamAdmin: true},
			},
			CanSee:  []int{2, 3, 4},
			CanEdit: []int{3, 4},
		},
	}

	// Insert all the teams
	testInsertTeam(t, repo, teams)

	// Get a team by its id
	for i, team := range teams {
		testGetTeam(t, repo, team.ID, team, fmt.Sprintf("get team %d", i))
	}

	// Get a team that does not exist
	testGetTeam(t, repo, 100, &Team{}, "team does not exist")

	// List all teams for a user
	testGetTeamsForUser(t, repo, 1, teams, "get teams for user")
	testGetTeamsForUser(t, repo, 100, []*Team{}, "get teams for user that does not exist")

	// Update a team's name
	teams[0].Name = "Pizza yolo"
	testUpdateTeam(t, repo, teams[0])
	testGetTeam(t, repo, teams[0].ID, teams[0], "get team 0 after name update")

	// Update a team's users (+2 -1)
	teams[1].Members = []TeamMember{
		{ID: 2, IsTeamAdmin: false},
		{ID: 3, IsTeamAdmin: false},
		{ID: 4, IsTeamAdmin: true},
	}
	testUpdateTeam(t, repo, teams[1])
	testGetTeam(t, repo, teams[1].ID, teams[1], "get team 1 after members update")

	// Update a team's permissions (canSee: +2 -1, canEdit: +2 -0)
	teams[1].CanSee = []int{1, 3, 4, 5}
	teams[1].CanEdit = []int{1, 3, 4, 5}
	testUpdateTeam(t, repo, teams[1])
	testGetTeam(t, repo, teams[1].ID, teams[1], "get team 1 after permissions update")

	// Delete a team
	testDeleteTeam(t, repo, teams[1].ID, "delete team 1")

	// List teams for user again
	testGetTeamsForUser(t, repo, 1, []*Team{teams[0]}, "get teams for user after delete")
}

func testInsertTeam(t *testing.T, repo TeamRepository, teams []*Team) {
	ids := make([]int, len(teams))
	for i, team := range teams {
		err := repo.Upsert(team)
		require.NoError(t, err, "insert %s should not fail", team.Name)
		require.NotEqual(t, 0, team.ID, "id should be set by insert")
		ids[i] = team.ID
	}

	// Test that all the ids are different
	sort.Ints(ids)
	for i := 0; i < len(ids)-1; i++ {
		require.NotEqual(t, ids[i], ids[i+1], "all ids should be different")
	}
}
func testGetTeam(t *testing.T, repo TeamRepository, id int, team *Team, name string) {
	retrieved, err := repo.Get(id)
	if assert.NoError(t, err, "get should not fail") {
		assertTeam(t, *team, retrieved, name)
	}
}

func testGetTeamsForUser(t *testing.T, repo TeamRepository, userID int, teams []*Team, name string) {
	retrieved, err := repo.GetForUser(userID)
	if !assert.NoError(t, err, "get for user should not fail") {
		return
	}

	if assert.Equal(t, len(teams), len(retrieved), "incorrect number of teams retrieved") {
		for i, team := range teams {
			assertTeam(t, *team, retrieved[i], name)
		}
	}
}

func testUpdateTeam(t *testing.T, repo TeamRepository, team *Team) {
	id := team.ID
	err := repo.Upsert(team)
	assert.NoError(t, err, "insert %s should not have failed", team.Name)
	assert.Equal(t, id, team.ID, "id should not change")
}

func testDeleteTeam(t *testing.T, repo TeamRepository, teamID int, name string) {
	err := repo.Delete(teamID)
	assert.NoError(t, err, "delete should not fail")

	retrieved, err := repo.Get(teamID)
	assert.NoError(t, err, "get after delete should not fail")
	assertTeam(t, Team{}, retrieved, name)
}

func assertTeam(t *testing.T, expected, actual Team, name string) {
	// General information
	assert.Equal(t, expected.ID, actual.ID, "%s - teams' ids should be equal", name)
	assert.Equal(t, expected.Name, actual.Name, "%s - teams' names should be equal", name)

	// Members
	if assert.Equal(t, len(expected.Members), len(actual.Members), "%s - number of members should be the same", name) {
		// Order does not matter
		for _, member := range expected.Members {
			assert.Contains(t, actual.Members, member, "%s - member %v should be in team", name, member)
		}
	}

	// Permissions
	if assert.Equal(t, len(expected.CanSee), len(actual.CanSee), "%s - number of seeable papers should be the same", name) {
		// Order does not matter
		for _, paperID := range expected.CanSee {
			assert.Contains(t, actual.CanSee, paperID, "%s - paperID %d should be in team's canSee", name, paperID)
		}
	}

	if assert.Equal(t, len(expected.CanEdit), len(actual.CanEdit), "%s - number of editable papers should be the same", name) {
		// Order does not matter
		for _, paperID := range expected.CanEdit {
			assert.Contains(t, actual.CanEdit, paperID, "%s - paperID %d should be in team's canEdit", name, paperID)
		}
	}
}
