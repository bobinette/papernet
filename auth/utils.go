package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"

	"github.com/bobinette/papernet/auth/jwt"
	"github.com/bobinette/papernet/errors"
)

// Variables and functions for specific errors
var (
	errInvalidRequest = errors.New("invalid request")
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

// HTTPServer defines the interface to register the http handlers.
type HTTPServer interface {
	RegisterHandler(path, method string, f http.Handler)
}

// statusCoder is useful to return http responses with a status that is not 200 but is not
// an error either.
type statusCoder struct {
	code int
}

func (s statusCoder) StatusCode() int { return s.code }

// extractUserID returns the user id present in the context, or an error if there is
// no user id or the claims are not correct.
func extractUserID(ctx context.Context) (int, error) {
	claims := ctx.Value(kitjwt.JWTClaimsContextKey)
	if claims == nil {
		return 0, errors.New("no user", errors.WithCode(http.StatusUnauthorized))
	}

	ppnClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return 0, errors.New("invalid claims", errors.WithCode(http.StatusForbidden))
	}

	return ppnClaims.UserID, nil
}

// encodeError writes an error as an HTTP response. It handles the status code
// contained in the error.
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	statusCode := http.StatusInternalServerError
	if err, ok := err.(errors.Error); ok {
		statusCode = err.Code()
	}
	w.WriteHeader(statusCode)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
