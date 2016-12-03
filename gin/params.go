package gin

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func queryBool(key string, c *gin.Context) (bool, bool, error) {
	v := c.Query(key)
	if v == "" {
		return false, false, nil
	}

	b, err := strconv.ParseBool(v)
	return b, true, err
}
