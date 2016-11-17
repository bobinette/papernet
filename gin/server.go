package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

func New(pr papernet.PaperRepository, ps papernet.PaperSearch, ts papernet.TagSearcher) (http.Handler, error) {
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
		c.Next()
	})

	// Unknown route
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})

	// Ping
	router.GET("/papernet/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"data": "ok"})
	})

	// Papers
	paperHandler := PaperHandler{Repository: pr, Searcher: ps, TagIndex: ts}
	paperHandler.RegisterRoutes(router)

	// Tags
	tagHandler := TagHandler{Searcher: ts}
	tagHandler.RegisterRoutes(router)

	return router, nil
}
