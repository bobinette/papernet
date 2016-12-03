package gin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
)

type HandlerFunc func(*gin.Context) (interface{}, error)

func JSONFormatter(next HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := next(c.Copy())
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

type Authenticator struct {
	Encoder        auth.Encoder
	UserRepository papernet.UserRepository
}

func (a *Authenticator) Authenticate(next HandlerFunc) HandlerFunc {
	return func(c *gin.Context) (interface{}, error) {
		token := c.Request.Header.Get("Authorization")
		if len(token) <= 6 || strings.ToLower(token[:7]) != "bearer " {
			return nil, errors.New("no token found")
		}

		userID, err := a.Encoder.Decode(token[7:])
		if err != nil {
			return nil, err
		}

		user, err := a.UserRepository.Get(userID)
		if err != nil {
			return nil, err
		}

		c.Set("user", user)
		return next(c)
	}
}
