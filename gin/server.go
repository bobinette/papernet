package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/errors"
)

type Server struct {
	*gin.Engine

	Authenticator auth.Authenticator
	Addr          string
}

func JSONRenderer(next papernet.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(map[string]string)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		req := papernet.Request{
			Request: c.Request,
			Params:  params,
		}
		res, err := next(&req)
		if err != nil {
			code := http.StatusInternalServerError
			if err, ok := err.(errors.Error); ok {
				code = err.Code()
			}

			c.JSON(code, map[string]interface{}{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func New(addr string, authenticator auth.Authenticator) (papernet.Server, error) {
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
		c.JSON(404, gin.H{"message": "Route not found"})
	})

	// Ping
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"data": "ok"})
	})

	return &Server{
		Engine:        router,
		Authenticator: authenticator,
		Addr:          addr,
	}, nil
}

func (s *Server) Register(route papernet.EndPoint) error {
	h := route.HandlerFunc
	if route.Authenticated {
		h = s.Authenticator.Authenticate(h)
	}
	s.Handle(route.Method, route.URL, JSONRenderer(h))
	return nil
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.Addr, s)
}
