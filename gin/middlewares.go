package gin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/errors"
)

type HandlerFunc func(*gin.Context) (interface{}, error)

func JSONFormatter(next HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := next(c.Copy())
		if err != nil {
			code := http.StatusInternalServerError
			if err, ok := err.(errors.Error); ok {
				code = err.Code()
			}

			c.JSON(code, map[string]interface{}{
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
			return nil, errors.New("no token found", errors.WithCode(http.StatusUnauthorized))
		}

		userID, err := a.Encoder.Decode(token[7:])
		if err != nil {
			return nil, errors.New("invalid token", errors.WithCode(http.StatusUnauthorized), errors.WithCause(err))
		}

		user, err := a.UserRepository.Get(userID)
		if err != nil {
			return nil, errors.New("could not get user", errors.WithCause(err))
		} else if user == nil {
			return nil, errors.New("unknown user", errors.WithCode(http.StatusUnauthorized))
		}

		c.Set("user", user)
		return next(c)
	}
}
