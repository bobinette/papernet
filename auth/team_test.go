package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bobinette/papernet/errors"
)

func TestTeamService(t *testing.T) {
	repo := NewInMemTeamRepository()
	userRepo := NewInMemUserRepository(repo)
	service := NewTeamService(repo, userRepo)

	// Insert users to be able to retrieve them by email when inviting
	admin := User{Email: "admin@paper.net", Owns: []int{1, 2}}
	require.NoError(t, userRepo.Upsert(&admin), "inserting admin must not fail")

	member := User{Email: "member@paper.net"}
	require.NoError(t, userRepo.Upsert(&member), "inserting member must not fail")

	otherMember := User{Email: "otherMember@paper.net"}
	require.NoError(t, userRepo.Upsert(&otherMember), "inserting otherMember must not fail")

	nonMember := User{Email: "nonMember@paper.net"}
	require.NoError(t, userRepo.Upsert(&nonMember), "inserting nonMember must not fail")

	// Pizza team
	team := Team{
		Name: "Pizza team",
		Members: []TeamMember{
			{ID: member.ID, IsTeamAdmin: true},
		},
		CanSee:  []int{1, 2},
		CanEdit: []int{3},
	}

	// Create a new team. The caller should be the only user of the team.
	// If the team has some, they are removed and the caller is set as admin.
	// If they are some permissions, they should be removed.
	createdTeam, err := service.Insert(admin.ID, team)
	require.NoError(t, err, "inserting must not fail")
	require.NotEqual(t, 0, createdTeam.ID, "created team should have an id")

	team = Team{
		ID:   createdTeam.ID,
		Name: "Pizza team",
		Members: []TeamMember{
			{ID: admin.ID, IsTeamAdmin: true},
		},
		CanSee:  []int{},
		CanEdit: []int{},
	}
	assertTeam(t, team, createdTeam, "team gotten from insert")

	otherTeam, err := service.Insert(member.ID, Team{Name: "Yolo team"})
	require.NoError(t, err, "inserting after delete must not fail")

	var retrieved Team

	// Invite a new member. If the user is not a member of the team, it should
	// get a 404, if not an admin -> 403, if the team does not exist -> 404
	retrieved, err = service.Invite(admin.ID, team.ID+1, member.Email)
	if assert.Error(t, err, "inviting in a non existing should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Invite(nonMember.ID, team.ID, member.Email)
	if assert.Error(t, err, "inviting from a non member should fail") {
		errors.AssertCode(t, err, 404)
	}

	team.Members = append(team.Members, TeamMember{ID: member.ID, IsTeamAdmin: false})
	retrieved, err = service.Invite(admin.ID, team.ID, member.Email)
	if assert.NoError(t, err, "inviting from an admin should not fail") {
		assertTeam(t, team, retrieved, "invited from admin")
	}

	retrieved, err = service.Invite(member.ID, team.ID, otherMember.Email)
	if assert.Error(t, err, "non admin member trying to invite should fail") {
		errors.AssertCode(t, err, 403)
	}

	retrieved, err = service.Invite(nonMember.ID, team.ID, member.Email)
	if assert.Error(t, err, "inviting in non existing should fail") {
		errors.AssertCode(t, err, 404)
	}

	team.Members = append(team.Members, TeamMember{ID: otherMember.ID, IsTeamAdmin: false})
	retrieved, err = service.Invite(admin.ID, team.ID, otherMember.Email)
	if assert.NoError(t, err, "inviting another member from an admin should not fail") {
		assertTeam(t, team, retrieved, "invited from admin again")
	}

	// Get the team. If the user is a member of the team, it should be good,
	// otherwise 404, and if the team does not exist -> 404
	retrieved, err = service.Get(admin.ID, team.ID)
	if assert.NoError(t, err, "getting team for admin should not fail") {
		assertTeam(t, team, retrieved, "retrieved for admin")
	}

	retrieved, err = service.Get(member.ID, team.ID)
	if assert.NoError(t, err, "getting team for member should not fail") {
		assertTeam(t, team, retrieved, "retrieved for member")
	}

	retrieved, err = service.Get(nonMember.ID, team.ID)
	if assert.Error(t, err, "getting team for non member should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Get(admin.ID, team.ID+1)
	if assert.Error(t, err, "getting non existing should fail") {
		errors.AssertCode(t, err, 404)
	}

	// Share a paper in a team.
	retrieved, err = service.Share(admin.ID, team.ID, 1, false)
	if assert.NoError(t, err, "admnin is member of team and owns 1, sharing should not fail") {
		assert.Contains(t, retrieved.CanSee, 1, "1 should be in team seeable papers")
		assert.NotContains(t, retrieved.CanEdit, 1, "1 should not be in team editable papers")
	}

	retrieved, err = service.Share(admin.ID, team.ID, 2, true)
	if assert.NoError(t, err, "admnin is member of team and owns 2, sharing with canEdit should not fail") {
		assert.Contains(t, retrieved.CanSee, 2, "2 should be in team seeable papers")
		assert.Contains(t, retrieved.CanEdit, 2, "2 should be in team editable papers")
	}

	retrieved, err = service.Share(nonMember.ID, team.ID, 4, true)
	if assert.Error(t, err, "nonMember is not in team, sharing should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Share(admin.ID, team.ID, 3, true)
	if assert.Error(t, err, "admin cannot see 3, sharing should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Share(member.ID, otherTeam.ID, 1, true)
	if assert.Error(t, err, "member can see 1 but is not the owner, sharing should fail") {
		errors.AssertCode(t, err, 403)
	}

	retrieved, err = service.Share(nonMember.ID, team.ID, 4, true)
	if assert.Error(t, err, "nonMember is not in team, sharing should fail") {
		errors.AssertCode(t, err, 404)
	}

	// Kick a member. If the user is not a member of the team, it should
	// get a 404, if the team does not exist -> 404 as well. If the user
	// is an admin, he/she can kick anyone that is not an admin. If the user
	// is not an admin, he/she can "kick" him/herself (i.e. leave the team)
	retrieved, err = service.Kick(admin.ID, team.ID+1, member.ID)
	if assert.Error(t, err, "kicking from a non existing team should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Kick(admin.ID, team.ID, nonMember.ID)
	if assert.Error(t, err, "kicking a user not in the team should fail") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Kick(nonMember.ID, team.ID, member.ID)
	if assert.Error(t, err, "non member should not be able to kick") {
		errors.AssertCode(t, err, 404)
	}

	retrieved, err = service.Kick(member.ID, team.ID, otherMember.ID)
	if assert.Error(t, err, "non admin member should not be able to kick") {
		errors.AssertCode(t, err, 403)
	}

	retrieved, err = service.Kick(member.ID, team.ID, member.ID)
	if assert.NoError(t, err, "member should be able to leave a team") {
		assert.NotContains(t, retrieved.Members, TeamMember{ID: member.ID, IsTeamAdmin: false}, "member should be in team anymore")
	}

	retrieved, err = service.Kick(admin.ID, team.ID, otherMember.ID)
	if assert.NoError(t, err, "admin should be able to kick member") {
		assert.NotContains(t, retrieved.Members, TeamMember{ID: otherMember.ID, IsTeamAdmin: false}, "member should be in team anymore")
	}

	// Invitations to test the get for user
	_, err = service.Invite(admin.ID, team.ID, member.Email)
	require.NoError(t, err, "inviting after delete in 1st team must not fail")
	_, err = service.Invite(member.ID, otherTeam.ID, otherMember.Email)
	require.NoError(t, err, "inviting after delete in 2nd team must not fail")

	// Get the users' teams. admin should have 1 team of which it is admin. member
	// should have 2 teams one of which it is admin. otherMember has 1 team of which
	// it is not admin. nonMember has no team.
	var teams []Team

	teams, err = service.GetForUser(admin.ID)
	assert.NoError(t, err, "get for admin should not fail")
	for _, retrieved = range teams {
		if retrieved.ID == team.ID {
			assert.True(t, userIsMemberOfTeam(admin.ID, retrieved), "admin should be member of the Pizza team")
			assert.True(t, userIsAdminOfTeam(admin.ID, retrieved), "admin should be admin of the Pizza team")
		} else {
			assert.Fail(t, fmt.Sprintf("team %d should not appear for admin", retrieved.ID))
		}
	}

	teams, err = service.GetForUser(member.ID)
	assert.NoError(t, err, "get for member should not fail")
	for _, retrieved = range teams {
		if retrieved.ID == team.ID {
			assert.True(t, userIsMemberOfTeam(member.ID, retrieved), "member should be member of the Pizza team")
			assert.False(t, userIsAdminOfTeam(member.ID, retrieved), "member should not be admin of the Pizza team")
		} else if retrieved.ID == otherTeam.ID {
			assert.True(t, userIsMemberOfTeam(member.ID, retrieved), "member should be member of the Pizza team")
			assert.True(t, userIsAdminOfTeam(member.ID, retrieved), "member should be admin of the Pizza team")
		} else {
			assert.Fail(t, fmt.Sprintf("team %d should not appear for member", retrieved.ID))
		}
	}

	teams, err = service.GetForUser(otherMember.ID)
	assert.NoError(t, err, "get for otherMember should not fail")
	for _, retrieved = range teams {
		if retrieved.ID == otherTeam.ID {
			assert.True(t, userIsMemberOfTeam(otherMember.ID, retrieved), "otherMember should be member of the Pizza team")
			assert.False(t, userIsAdminOfTeam(otherMember.ID, team), "otherMember should not be admin of the Pizza team")
		} else {
			assert.Fail(t, fmt.Sprintf("team %d should not appear for otherMember", retrieved.ID))
		}
	}

	// Delete a team. Only the admin should be able to delete a team.
	// Non member -> 404. Non admin -> 403.
	err = service.Delete(nonMember.ID, createdTeam.ID)
	if assert.Error(t, err, "deleting from a non member team should fail") {
		errors.AssertCode(t, err, 404)
	}

	err = service.Delete(member.ID, createdTeam.ID)
	if assert.Error(t, err, "deleting from a non admin team should fail") {
		errors.AssertCode(t, err, 403)
	}

	err = service.Delete(admin.ID, createdTeam.ID)
	assert.NoError(t, err, "deleting from an admin should be ok")
}

func TestUserIsMemberOfTeam(t *testing.T) {
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
		"non admin member": {1, true},
		"admin member":     {3, true},
		"not a member":     {5, false},
	}

	for name, tt := range tts {
		isAdmin := userIsMemberOfTeam(tt.userID, team)
		assert.Equal(t, tt.isAdmin, isAdmin, name)
	}
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
