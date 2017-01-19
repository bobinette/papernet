package web

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/errors"
)

func getUser(ctx context.Context) (*papernet.User, error) {
	u := ctx.Value("user")
	if u == nil {
		return nil, errors.New("could not extract user", errors.WithCode(http.StatusUnauthorized))
	}

	user, ok := u.(*papernet.User)
	if !ok {
		return nil, errors.New("could not retrieve user", errors.WithCode(http.StatusUnauthorized))
	}

	return user, nil
}

func queryBool(key string, req papernet.Request) (bool, bool, error) {
	v := req.Query(key)
	if v == "" {
		return false, false, nil
	}

	b, err := strconv.ParseBool(v)
	return b, true, err
}

func queryInt(key string, req papernet.Request) (int, bool, error) {
	v := req.Query(key)
	if v == "" {
		return 0, false, nil
	}

	i, err := strconv.Atoi(v)
	return i, true, err
}

func queryUInt64(key string, req papernet.Request) (uint64, bool, error) {
	v := req.Query(key)
	if v == "" {
		return 0, false, nil
	}

	i, err := strconv.ParseUint(v, 10, 64)
	return i, true, err
}
