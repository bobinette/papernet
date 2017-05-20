package services

import (
	"net/url"
	"sync"
	"testing"

	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleService_LoginURL(t *testing.T) {
	redirectURI := "http://redirect-url.com"
	clientID := "client_id"
	responseType := "code"

	service := &GoogleService{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "",
			RedirectURL:  redirectURI,
			Scopes:       scopes,
			Endpoint:     googleEndpoint,
		},

		stateMutex: &sync.RWMutex{},
		state:      make(map[string]struct{}),
	}

	loginURLString := service.LoginURL()
	loginURL, err := url.Parse(loginURLString)
	require.NoError(t, err, "url should be valid")

	// Assert scheme and host
	assert.Equal(t, "https", loginURL.Scheme, "scheme should be https")
	assert.Equal(t, "accounts.google.com", loginURL.Host, "host should be google")

	// Assert query parameters
	query := loginURL.Query()
	assert.Equal(t, scopes, query["scope"], "invalid scope")
	assert.Equal(t, redirectURI, query.Get("redirect_uri"), "invalid redirect uri")
	assert.Equal(t, clientID, query.Get("client_id"), "invalid client id")
	assert.Equal(t, responseType, query.Get("response_type"), "invalid response type")
	assert.NotEqual(t, "", query.Get("state"), "state should not be empty")
	assert.Contains(t, service.state, query.Get("state"), "state be stored in service")
}
