package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertCode(t *testing.T, err error, code int) {
	switch err := err.(type) {
	case Error:
		assert.Equal(t, code, err.Code(), "code should be equal")
	default:
		if code != DefaultCode {
			assert.Fail(t, fmt.Sprintf("error is not Error and expected code != %d (default)", DefaultCode))
		}
	}
}
