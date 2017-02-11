package web

import (
	"fmt"
	"strconv"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/errors"
)

type Request struct {
	*papernet.Request
}

func WrapRequest(next func(*Request) (interface{}, error)) papernet.HandlerFunc {
	return func(req *papernet.Request) (interface{}, error) {
		return next(&Request{req})
	}
}

func (r *Request) Query(key string, v interface{}) error {
	q := r.Request.Query(key)

	// If q is empty, let v has its default value
	if q == "" {
		return nil
	}

	// Type switch to fill v
	switch v := v.(type) {
	case *string:
		*v = q
	case *bool:
		b, err := strconv.ParseBool(q)
		if err != nil {
			return errors.New(fmt.Sprintf("could not parse %s", q), errors.WithCause(err))
		}
		*v = b
	case *int:
		i, err := strconv.Atoi(q)
		if err != nil {
			return errors.New(fmt.Sprintf("could not parse %s", q), errors.WithCause(err))
		}
		*v = i
	case *uint64:
		i, err := strconv.ParseUint(q, 10, 64)
		if err != nil {
			return errors.New(fmt.Sprintf("could not parse %s", q), errors.WithCause(err))
		}
		*v = i
	default:
		return errors.New(fmt.Sprintf("unsupported type: %T", v))
	}

	return nil
}
