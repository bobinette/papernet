package gin

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/errors"
)

func GetUser(c *gin.Context) (*papernet.User, error) {
	u, ok := c.Get("user")
	if !ok {
		return nil, errors.New("could not extract user", errors.WithCode(http.StatusUnauthorized))
	}

	user, ok := u.(*papernet.User)
	if !ok {
		return nil, errors.New("could not retrieve user", errors.WithCode(http.StatusUnauthorized))
	}

	return user, nil
}

type UserHandler struct {
	GoogleClient  *auth.GoogleClient
	Repository    papernet.UserRepository
	Authenticator Authenticator
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/api/auth", JSONFormatter(h.AuthURL))
	router.GET("/api/auth/google", JSONFormatter(h.Google))

	router.GET("/api/me", JSONFormatter(h.Authenticator.Authenticate(h.Me)))
	router.POST("/api/bookmarks", JSONFormatter(h.Authenticator.Authenticate(h.UpdateBookmarks)))
}

func (h *UserHandler) AuthURL(c *gin.Context) (interface{}, error) {
	return map[string]interface{}{
		"url": h.GoogleClient.LoginURL(),
	}, nil
}

func (h *UserHandler) Google(c *gin.Context) (interface{}, error) {
	state := c.Query("state")
	code := c.Query("code")

	user, err := h.GoogleClient.ExchangeToken(state, code)
	if err != nil {
		return nil, errors.New("error exchanging token", errors.WithCause(err))
	}

	if dbUser, err := h.Repository.Get(user.ID); err != nil {
		return nil, errors.New("error checking user in db", errors.WithCause(err))
	} else if dbUser == nil {
		err = h.Repository.Upsert(user)
		if err != nil {
			return nil, errors.New("error saving user", errors.WithCause(err))
		}
	} else {
		user = dbUser
	}

	token, err := h.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		return nil, errors.New("error encoding token", errors.WithCause(err))
	}

	return map[string]interface{}{
		"access_token": token,
	}, nil
}

func (h *UserHandler) Me(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": user,
	}, nil
}

func (h *UserHandler) UpdateBookmarks(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Add    []int `json:"add"`
		Remove []int `json:"remove"`
	}
	err = c.BindJSON(&payload)
	if err != nil {
		return nil, errors.New("error reading body", errors.WithCode(http.StatusBadRequest), errors.WithCause(err))
	}

	bookmarks := make(map[int]struct{})
	for _, b := range user.Bookmarks {
		bookmarks[b] = struct{}{}
	}
	for _, b := range payload.Add {
		bookmarks[b] = struct{}{}
	}
	for _, b := range payload.Remove {
		delete(bookmarks, b)
	}

	user.Bookmarks = func(m map[int]struct{}) []int {
		i := 0
		a := make([]int, len(m))
		for k, _ := range m {
			a[i] = k
			i++
		}
		sort.Ints(a)
		return a
	}(bookmarks)

	err = h.Repository.Upsert(user)
	if err != nil {
		return nil, errors.New("error saving", errors.WithCause(err))
	}

	return map[string]interface{}{
		"data": user,
	}, nil
}
