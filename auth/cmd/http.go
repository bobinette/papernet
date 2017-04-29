package auth

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/log"
)

type Configuration struct {
	KeyPath string `toml:"key"`
	Google  string `toml:"google"`

	Cayley struct {
		Store string `toml:"store"`
	} `toml:"cayley"`
}

// Start registers
func Start(srv auth.HTTPServer, conf Configuration, logger log.Logger) *auth.UserService {
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
	store, err := cayley.NewStore(conf.Cayley.Store)
	if err != nil {
		logger.Fatal("could not create user graph:", err)
	}
	userRepository := cayley.NewUserRepository(store)
	teamRepository := cayley.NewTeamRepository(store)

	// Start user endpoint
	userService := auth.NewUserService(userRepository)
	auth.RegisterUserHTTP(srv, userService, []byte(key.Key))

	// Start team endpoint
	teamService := auth.NewTeamService(teamRepository, userRepository)
	auth.RegisterTeamHTTP(srv, teamService, []byte(key.Key))

	return userService
}
