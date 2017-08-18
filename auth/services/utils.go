package services

import (
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/bobinette/papernet/errors"
)

// errUserNotFound returns a 404 for when a user could not be found.
func errUserNotFound(id int) error {
	return errors.New(fmt.Sprintf("No user for id %d", id), errors.NotFound())
}

// errPaperNotFound returns a 404 for when a paper could not be found.
func errPaperNotFound(id int) error {
	return errors.New(fmt.Sprintf("No paper for id %d", id), errors.NotFound())
}

// errTeamNotFound returns a 404 for when a team could not be found.
func errTeamNotFound(id int) error {
	return errors.New(fmt.Sprintf("No team for id %d", id), errors.NotFound())
}

// errNotTeamAdmin returns a 403 for when team admin privilege is needed
func errNotTeamAdmin(id int) error {
	return errors.New(fmt.Sprintf("You are not an admin of team %d", id), errors.Forbidden())
}

func randToken(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
