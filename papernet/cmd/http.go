package paper

import (
	"encoding/json"
	"io/ioutil"

	// "github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/papernet/bleve"
	"github.com/bobinette/papernet/papernet/bolt"
	"github.com/bobinette/papernet/papernet/http"
	"github.com/bobinette/papernet/papernet/services"

	authClient "github.com/bobinette/papernet/clients/auth"
)

type Configuration struct {
	KeyPath string `toml:"key"`
	Bleve   struct {
		Store string `toml:"store"`
	} `toml:"bleve"`
	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

// Start registers
func Start(srv http.Server, conf Configuration, logger log.Logger, au *authClient.Client) *services.PaperService {
	// Load key from file
	keyData, err := ioutil.ReadFile(conf.KeyPath)
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

	// Create repositories
	boltDriver := bolt.Driver{}
	err = boltDriver.Open(conf.Bolt.Store)
	if err != nil {
		logger.Fatalf("could not open bolt: %v", err)
	}
	paperRepository := bolt.PaperRepository{Driver: &boltDriver}
	tagIndex := bolt.TagIndex{Driver: &boltDriver}

	// Create index
	index := bleve.PaperIndex{}
	err = index.Open(conf.Bleve.Store)
	if err != nil {
		logger.Fatalf("could not open bleve: %v", err)
	}

	// Create services
	tagService := services.NewTagService(&tagIndex)
	paperService := services.NewPaperService(&paperRepository, &index, au, tagService)

	// Register paper endpoints
	http.RegisterPaperEndpoints(srv, paperService, []byte(key.Key), au)
	http.RegisterTagEndpoints(srv, tagService, []byte(key.Key), au)

	return paperService
}
