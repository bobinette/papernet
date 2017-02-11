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

	Authenticator Authenticator
}

func GetUser(c *gin.Context) (*papernet.User, error) {
	u, ok := c.Get("user")
	if !ok {
		return nil, errors.New("could not extract user", errors.WithCode(http.StatusUnauthorized))
	}

	user, ok := u.(*papernet.User)
	if !ok {
		return nil, errors.New("could not retrieve user", errors.WithCode(http.StatusUnauthorized))
	}

	return user, nil
}

func New(
	ps papernet.PaperStore,
	pi papernet.PaperIndex,
	ts papernet.TagIndex,
	ur papernet.UserRepository,
	sk papernet.SigningKey,
	googleOAuthClient *auth.GoogleClient,
) (papernet.Server, error) {
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

	encoder := auth.EncodeDecoder{Key: sk.Key}
	authenticator := Authenticator{Encoder: encoder, UserRepository: ur}

	// Tags
	tagHandler := TagHandler{Searcher: ts}
	tagHandler.RegisterRoutes(router)

	return &Server{
		router,
		authenticator,
	}, nil
}

func (s *Server) Register(route papernet.Route) error {
	h := route.HandlerFunc
	if route.Authenticated {
		h = s.Authenticator.AuthenticateP(h)
	}
	s.Handle(route.Method, route.Route, JSONRenderer(h))
	return nil
}
