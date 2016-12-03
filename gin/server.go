package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/oauth"
)

func New(
	pr papernet.PaperRepository,
	ps papernet.PaperIndex,
	ts papernet.TagIndex,
	ur papernet.UserRepository,
	sk papernet.SigningKey,
	googleOAuthClient *oauth.GoogleOAuthClient,
) (http.Handler, error) {
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept-Language, Authorization, Content-Type")
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

	// Auth
	authHandler := AuthHandler{GoogleClient: googleOAuthClient, Repository: ur, SigningKey: sk}
	authHandler.RegisterRoutes(router)

	return router, nil
}
