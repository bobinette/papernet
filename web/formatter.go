package web

import (
	"github.com/gin-gonic/gin"
)

type Formatter struct{}

func (f *Formatter) Wrap(next func(*gin.Context) (interface{}, int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, code, err := next(c)

		if err != nil {
			c.JSON(code, map[string]interface{}{
				"data":    "ko",
				"message": err.Error(),
			})
			return
		}

		c.JSON(code, map[string]interface{}{
			"data": data,
		})
	}
}
