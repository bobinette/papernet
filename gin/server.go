package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
)

func New(
	ps papernet.PaperStore,
	pi papernet.PaperIndex,
	ts papernet.TagIndex,
	ur papernet.UserRepository,
	sk papernet.SigningKey,
	googleOAuthClient *auth.GoogleClient,
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
	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"data": "ok"})
	})

	encoder := auth.Encoder{Key: sk.Key}
	authenticator := Authenticator{Encoder: encoder, UserRepository: ur}

	// Papers
	paperHandler := PaperHandler{
		Store:          ps,
		Searcher:       pi,
		UserRepository: ur,
		TagIndex:       ts,
		Authenticator:  authenticator,
	}
	paperHandler.RegisterRoutes(router)

	// Tags
	tagHandler := TagHandler{Searcher: ts}
	tagHandler.RegisterRoutes(router)

	// Auth
	userHandler := UserHandler{GoogleClient: googleOAuthClient, Repository: ur, Authenticator: authenticator}
	userHandler.RegisterRoutes(router)

	// Arxiv
	arxivHandler := ArxivHandler{Authenticator: authenticator}
	arxivHandler.RegisterRoutes(router)

	return router, nil
}
