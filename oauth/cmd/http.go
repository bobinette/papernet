package oauth

import (
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/oauth"
	"github.com/bobinette/papernet/oauth/http"
	"github.com/bobinette/papernet/oauth/services"
)

type Configuration struct {
	Provider   string `toml:"provider"`
	GooglePath string `toml:"google"`
}

func Start(srv http.Server, cfg Configuration, logger log.Logger, userService oauth.UserService) {
	// OAuth service
	authUserService := oauth.NewUserClient(userService)
	googleService, err := services.NewGoogleService(cfg.GooglePath, authUserService)
	if err != nil {
		logger.Fatal("could not instantiate google service", err)
	}
	http.RegisterGoogleHTTPRoutes(srv, googleService)
}
