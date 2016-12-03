package errors

import (
	"fmt"
)

type Error interface {
	error

	Code() int
	Message() string
	Cause() error
}

// Default code defines the code that will be used by default when
// none is given. It is set to 500, Internal Server Error
var DefaultCode = 500

type myError struct {
	code  int
	msg   string
	cause *myError
}

func (err *myError) Error() string {
	if err.cause == nil {
		return err.msg
	}

	return fmt.Sprintf("%s: %v", err.msg, err.cause)
}

func (err *myError) Code() int {
	return err.code
}

func (err *myError) Message() string {
	return err.msg
}

func (err *myError) Cause() error {
	return err.cause
}

type ErrorEnricher func(error) error

func WithCode(code int) func(error) error {
	return func(err error) error {
		switch err := err.(type) {
		case *myError:
			err.code = code
			return err
		}

		// default
		return &myError{
			msg:   err.Error(),
			code:  code,
			cause: nil,
		}
	}
}

func WithCause(cause error) func(error) error {
	var myCause *myError
	switch cause := cause.(type) {
	case *myError:
		myCause = cause
	default:
		myCause = &myError{msg: cause.Error(), code: DefaultCode, cause: nil}
	}

	return func(err error) error {
		if myErr, ok := err.(*myError); ok {
			myErr.cause = myCause
			return myErr
		}

		return &myError{
			msg:   err.Error(),
			code:  myCause.code,
			cause: myCause,
		}
	}
}

func New(msg string, fs ...ErrorEnricher) error {
	var err error
	err = &myError{
		msg:   msg,
		code:  DefaultCode,
		cause: nil,
	}

	for _, f := range fs {
		err = f(err)
	}

	return err
}
