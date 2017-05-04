package endpoints

import (
	"context"
	"net/http"

	kitjwt "github.com/go-kit/kit/auth/jwt"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/jwt"
)

// Variables and functions for specific errors
var (
	errInvalidRequest = errors.New("invalid request")
)

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
