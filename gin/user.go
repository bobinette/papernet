package gin

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
)

func GetUser(c *gin.Context) (*papernet.User, error) {
	u, ok := c.Get("user")
	if !ok {
		return nil, errors.New("could not extract user")
	}

	user, ok := u.(*papernet.User)
	if !ok {
		return nil, errors.New("could not extract user")
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

func (h *UserHandler) Me(c *gin.Context) (interface{}, error) {
	user, err := GetUser(c)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": user,
	}, nil
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
		return nil, err
	}

	err = h.Repository.Upsert(user)
	if err != nil {
		return nil, err
	}

	token, err := h.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token": token,
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
		return nil, err
	}

	in := func(i int, a []int) bool {
		for _, e := range a {
			if e == i {
				return true
			}
		}
		return false
	}

	bookmarks := make(map[int]struct{})
	for _, b := range user.Bookmarks {
		if !in(b, payload.Remove) {
			bookmarks[b] = struct{}{}
		}
	}

	for _, b := range payload.Add {
		if !in(b, payload.Remove) {
			bookmarks[b] = struct{}{}
		}
	}

	user.Bookmarks = func(m map[int]struct{}) []int {
		i := 0
		a := make([]int, len(m))
		for k, _ := range m {
			a[i] = k
			i++
		}
		return a
	}(bookmarks)
	log.Println(user.Bookmarks)

	err = h.Repository.Upsert(user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user": user,
	}, nil
}
