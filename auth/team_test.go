package auth

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bobinette/papernet/errors"
)

func TestTeamService(t *testing.T) {
	repo := NewInMemTeamRepository()
	service := NewTeamService(repo)

	adminID := 1
	memberID := 2
	otherMemberID := 3
	nonMemberID := 4

	team := Team{
		Name: "Pizza team",
		Members: []TeamMember{
			{ID: memberID, IsTeamAdmin: true},
		},
		CanSee:  []int{1, 2},
		CanEdit: []int{3},
	}

	// Create a new team. The caller should be the only user of the team.
	// If the team has some, they are removed and the caller is set as admin.
	// If they are some permissions, they should be removed.
	createdTeam, err := service.Insert(adminID, team)
	require.NoError(t, err, "inserting must not fail")
	require.NotEqual(t, 0, createdTeam.ID, "created team should have an id")

	team = Team{
		ID:   createdTeam.ID,
		Name: "Pizza team",
		Members: []TeamMember{
			{ID: adminID, IsTeamAdmin: true},
		},
		CanSee:  []int{},
		CanEdit: []int{},
	}
	assertTeam(t, team, createdTeam, "team gotten from insert")

	var retrieved Team

	// Invite a new member. If the user is not a member of the team, it should
	// get a 404, if not an admin -> 403, if the team does not exist -> 404
	retrieved, err = service.Invite(adminID, team.ID+1, memberID)
	if assert.Error(t, err, "inviting in a non existing should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Invite(nonMemberID, team.ID, memberID)
	if assert.Error(t, err, "inviting from a non member should fail") {
		errors.AssertCode(t, err, 404)
	}

	team.Members = append(team.Members, TeamMember{ID: memberID, IsTeamAdmin: false})
	retrieved, err = service.Invite(adminID, team.ID, memberID)
	if assert.NoError(t, err, "inviting from an admin should not fail") {
		assertTeam(t, team, retrieved, "invited from admin")
	}

	retrieved, err = service.Invite(memberID, team.ID, otherMemberID)
	if assert.Error(t, err, "non admin member trying to invite should fail") {
		errors.AssertCode(t, err, 403)
	}

	retrieved, err = service.Invite(nonMemberID, team.ID, memberID)
	if assert.Error(t, err, "inviting in non existing should fail") {
		errors.AssertCode(t, err, 404)
	}

	team.Members = append(team.Members, TeamMember{ID: otherMemberID, IsTeamAdmin: false})
	retrieved, err = service.Invite(adminID, team.ID, otherMemberID)
	if assert.NoError(t, err, "inviting another member from an admin should not fail") {
		assertTeam(t, team, retrieved, "invited from admin again")
	}

	// Get the team. If the user is a member of the team, it should be good,
	// otherwise 404, and if the team does not exist -> 404
	retrieved, err = service.Get(adminID, team.ID)
	if assert.NoError(t, err, "getting team for admin should not fail") {
		assertTeam(t, team, retrieved, "retrieved for admin")
	}

	retrieved, err = service.Get(memberID, team.ID)
	if assert.NoError(t, err, "getting team for member should not fail") {
		assertTeam(t, team, retrieved, "retrieved for member")
	}

	retrieved, err = service.Get(nonMemberID, team.ID)
	if assert.Error(t, err, "getting team for non member should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Get(adminID, team.ID+1)
	if assert.Error(t, err, "getting non existing should fail") {
		errors.AssertCode(t, err, 404)
	}

	// Kick a member. If the user is not a member of the team, it should
	// get a 404, if the team does not exist -> 404 as well. If the user
	// is an admin, he/she can kick anyone that is not an admin. If the user
	// is not an admin, he/she can "kick" him/herself (i.e. leave the team)
	retrieved, err = service.Kick(adminID, team.ID+1, memberID)
	if assert.Error(t, err, "kicking from a non existing team should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Kick(adminID, team.ID, nonMemberID)
	if assert.Error(t, err, "kicking a user not in the team should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Kick(nonMemberID, team.ID, memberID)
	if assert.Error(t, err, "non member should not be able to kick") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Kick(memberID, team.ID, otherMemberID)
	if assert.Error(t, err, "non admin member should not be able to kick") {
		errors.AssertCode(t, err, 403)
	}

	retrieved, err = service.Kick(memberID, team.ID, memberID)
	if assert.NoError(t, err, "member should be able to leave a team") {
		assert.NotContains(t, retrieved.Members, TeamMember{ID: memberID, IsTeamAdmin: false}, "member should be in team anymore")
	}

	retrieved, err = service.Kick(adminID, team.ID, otherMemberID)
	if assert.NoError(t, err, "admin should be able to kick member") {
		assert.NotContains(t, retrieved.Members, TeamMember{ID: otherMemberID, IsTeamAdmin: false}, "member should be in team anymore")
	}

	// Delete a team. Only the admin should be able to delete a team.
	// Non member -> 404. Non admin -> 403.
	createdTeam, err = service.Insert(adminID, team)
	require.NoError(t, err, "inserting (for delete) must not fail")
	_, err = service.Invite(adminID, createdTeam.ID, memberID)
	require.NoError(t, err, "inviting (for delete) must not fail")

	err = service.Delete(nonMemberID, createdTeam.ID)
	if assert.Error(t, err, "deleting from a non member team should fail") {
		errors.AssertCode(t, err, 404)
	}

	err = service.Delete(memberID, createdTeam.ID)
	if assert.Error(t, err, "deleting from a non admin team should fail") {
		errors.AssertCode(t, err, 403)
	}

	err = service.Delete(adminID, createdTeam.ID)
	assert.NoError(t, err, "deleting from an admin should be ok")
}

func TestUserIsAdminOfTeam(t *testing.T) {
	team := Team{
		Members: []TeamMember{
			{ID: 1, IsTeamAdmin: false},
			{ID: 2, IsTeamAdmin: false},
			{ID: 3, IsTeamAdmin: true},
			{ID: 4, IsTeamAdmin: false},
		},
	}
	tts := map[string]struct {
		userID  int
		isAdmin bool
	}{
		"non admin member": {1, false},
		"admin member":     {3, true},
		"not a member":     {5, false},
	}

	for name, tt := range tts {
		isAdmin := userIsAdminOfTeam(tt.userID, team)
		assert.Equal(t, tt.isAdmin, isAdmin, name)
	}
}
