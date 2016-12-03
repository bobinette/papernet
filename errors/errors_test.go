package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestWithCode(t *testing.T) {
	tts := []struct {
		err      error
		code     int
		expected *myError
	}{
		{
			err:  errors.New("simple error"),
			code: 404,
			expected: &myError{
				msg:   "simple error",
				code:  404,
				cause: nil,
			},
		},
		{
			err: &myError{
				msg:   "custom error",
				code:  200,
				cause: nil,
			},
			code: 501,
			expected: &myError{
				msg:   "custom error",
				code:  501,
				cause: nil,
			},
		},
		{
			err: &myError{
				msg:   "keep cause",
				code:  125,
				cause: &myError{msg: "I am the cause"},
			},
			code: 305,
			expected: &myError{
				msg:   "keep cause",
				code:  305,
				cause: &myError{msg: "I am the cause"},
			},
		},
		{
			// nil input should give nil output
			err:      nil,
			code:     305,
			expected: nil,
		},
	}

	for i, tt := range tts {
		err, _ := WithCode(tt.code)(tt.err).(*myError)
		assertErrors(tt.expected, err, t, fmt.Sprintf("%d WithCode", i))
	}
}

func TestWithCause(t *testing.T) {
	tts := []struct {
		err      error
		cause    error
		expected *myError
	}{
		{
			err:   errors.New("simple error"),
			cause: errors.New("I am the cause"),
			expected: &myError{
				msg:   "simple error",
				code:  500,
				cause: &myError{msg: "I am the cause", code: DefaultCode, cause: nil},
			},
		},
		{
			err: errors.New("simple error"),
			cause: &myError{
				msg:   "forward code",
				code:  120,
				cause: nil,
			},
			expected: &myError{
				msg:   "simple error",
				code:  120,
				cause: &myError{msg: "forward code", code: 120, cause: nil},
			},
		},
		{
			err: &myError{
				msg:   "custom error",
				code:  200,
				cause: nil,
			},
			cause: &myError{
				msg:   "custom cause",
				code:  300,
				cause: nil,
			},
			expected: &myError{
				msg:   "custom error",
				code:  200,
				cause: &myError{msg: "custom cause", code: 300, cause: nil},
			},
		},
		{
			err: &myError{
				msg:   "change cause",
				code:  125,
				cause: &myError{msg: "I am the cause", code: DefaultCode, cause: nil},
			},
			cause: errors.New("I am the new cause"),
			expected: &myError{
				msg:   "change cause",
				code:  125,
				cause: &myError{msg: "I am the new cause", code: DefaultCode, cause: nil},
			},
		},
		{
			// nil input should give nil output
			err:      nil,
			cause:    errors.New("The cause is ignored if the wrapper is nil"),
			expected: nil,
		},
	}

	for i, tt := range tts {
		err, _ := WithCause(tt.cause)(tt.err).(*myError)
		assertErrors(tt.expected, err, t, fmt.Sprintf("%d WithClause", i))
	}
}

func assertErrors(exp *myError, got *myError, t *testing.T, name string) {
	if exp == nil && got == nil {
		return
	}

	if exp == nil && got != nil {
		t.Errorf("%s - expected nil, got non-nil", name)
		return
	}

	if exp != nil && got == nil {
		t.Errorf("%s - expected non-nil, got nil", name)
		return
	}

	if got.code != exp.code {
		t.Errorf("%s - code: %d != %d", name, exp.code, got.code)
	}

	if got.msg != exp.msg {
		t.Errorf("%s - msg: %s != %s", name, exp.msg, got.msg)
	}

	assertErrors(exp.cause, got.cause, t, name)
}
