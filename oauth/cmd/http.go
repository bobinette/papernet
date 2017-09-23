package oauth

import (
	"net/http"

	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/oauth"
)

type Configuration struct {
	Email  bool `toml:"email"`
	Google bool `toml:"google"`
}

// Server defines the interface to register the http handlers.
type Server interface {
	RegisterHandler(path, method string, f http.Handler)
}

func Start(srv Server, cfg Configuration, logger log.Logger) {
	providerService := oauth.NewProviderService()

	// Basic email / password
	if cfg.Email {
		providerService.Register("email")
	}

	// Google
	if cfg.Google {
		providerService.Register("google")
	}

	oauth.RegisterProviderHTTPRoutes(srv, providerService)
}
