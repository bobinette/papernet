package oauth

import (
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/oauth"
	"github.com/bobinette/papernet/oauth/bolt"
	"github.com/bobinette/papernet/oauth/http"
	"github.com/bobinette/papernet/oauth/services"
)

type Configuration struct {
	Provider string `toml:"provider"`
	Bolt     string `toml:"bolt"`
	Email    struct {
		Enabled bool `toml:"enabled"`
	} `toml:"email"`
	Google struct {
		Enabled bool   `toml:"enabled"`
		File    string `toml:"file"`
	} `toml:"google"`
}

func Start(srv http.Server, cfg Configuration, logger log.Logger, userService oauth.UserService) {
	userClient := oauth.NewUserClient(userService)
	providerService := services.NewProviderService()

	boltDriver := &bolt.Driver{}
	if err := boltDriver.Open(cfg.Bolt); err != nil {
		logger.Fatal("could not open bolt driver", err)
	}

	// Basic email / password
	if cfg.Email.Enabled {
		repository := bolt.NewEmailRepository(boltDriver)
		service := services.NewEmailService(repository, userClient)
		http.RegisterEmailHTTPRoutes(srv, service)
		providerService.Register("email")
	}

	// Google
	if cfg.Google.Enabled {
		repository := bolt.NewGoogleRepository(boltDriver)
		service, err := services.NewGoogleService(repository, cfg.Google.File, userClient)
		if err != nil {
			logger.Fatal("could not instantiate google service", err)
		}
		http.RegisterGoogleHTTPRoutes(srv, service)
		providerService.Register("google")
	}

	http.RegisterProviderHTTPRoutes(srv, providerService)
}
