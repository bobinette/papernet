package oauth

import (
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/clients/auth"

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

func Start(srv http.Server, cfg Configuration, logger log.Logger, authClient *auth.Client) {
	providerService := services.NewProviderService()

	boltDriver := &bolt.Driver{}
	if err := boltDriver.Open(cfg.Bolt); err != nil {
		logger.Fatal("could not open bolt driver", err)
	}

	// Basic email / password
	if cfg.Email.Enabled {
		providerService.Register("email")
	}

	// Google
	if cfg.Google.Enabled {
		repository := bolt.NewGoogleRepository(boltDriver)
		service, err := services.NewGoogleService(repository, cfg.Google.File, authClient)
		if err != nil {
			logger.Fatal("could not instantiate google service", err)
		}
		http.RegisterGoogleHTTPRoutes(srv, service)
		providerService.Register("google")
	}

	http.RegisterProviderHTTPRoutes(srv, providerService)
}
