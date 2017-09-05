package oauth

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/clients/auth"

	"github.com/bobinette/papernet/google"
	"github.com/bobinette/papernet/google/bolt"
)

type Configuration struct {
	KeyPath string `toml:"key"`
	Bolt    string `toml:"bolt"`
	File    string `toml:"file"`
}

func Start(srv google.Server, cfg Configuration, logger log.Logger, authClient *auth.Client) {
	// Load key from file
	keyData, err := ioutil.ReadFile(cfg.KeyPath)
	if err != nil {
		logger.Fatal("could not open key file:", err)
	}

	// Extract key from data
	var key struct {
		Key string `json:"k"`
	}
	err = json.Unmarshal(keyData, &key)
	if err != nil {
		logger.Fatal("could not read key file:", err)
	}

	// Create service
	boltDriver := &bolt.Driver{}
	if err := boltDriver.Open(cfg.Bolt); err != nil {
		logger.Fatal("could not open bolt driver", err)
	}

	repository := bolt.NewUserRepository(boltDriver)
	userClient := google.NewUserClient(authClient)
	service, err := google.NewService(repository, cfg.File, userClient)
	if err != nil {
		logger.Fatal("could not instantiate google service", err)
	}
	google.RegisterGoogleHTTPRoutes(srv, service, []byte(key.Key), authClient)
}
