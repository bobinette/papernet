package web

import (
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

func NewServer(dbPath, indexPath string) (http.Handler, error) {
	store, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	ph, err := NewPaperHandler(store, indexPath)
	if err != nil {
		return nil, err
	}

	uh, err := NewUserHandler(store)
	if err != nil {
		return nil, err
	}

	uptime := UptimeHandler{f: Formatter{}}

	// Create router
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
		c.Next()
	})

	// Register routes
	uptime.Register(router)
	ph.Register(router)
	uh.Register(router)

	router.Static("/app", "./app")

	// Basic response in JSON if route not found
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})

	return router, nil
}
