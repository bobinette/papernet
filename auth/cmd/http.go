package auth

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/auth/http"
	"github.com/bobinette/papernet/auth/services"
	"github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/log"
)

type Configuration struct {
	KeyPath string `toml:"key"`
	Cayley  struct {
		Store string `toml:"store"`
	} `toml:"cayley"`
}

// Start registers
func Start(srv http.Server, conf Configuration, logger log.Logger) *services.UserService {
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
	tokenEncoder := jwt.NewEncoder([]byte(key.Key))

	// Create repositories
	store, err := cayley.NewStore(conf.Cayley.Store)
	if err != nil {
		logger.Fatal("could not create user graph:", err)
	}
	userRepository := cayley.NewUserRepository(store)
	teamRepository := cayley.NewTeamRepository(store)

	// Start user endpoint
	userService := services.NewUserService(userRepository, tokenEncoder)
	http.RegisterUserEndpoints(srv, userService, []byte(key.Key))

	// Start team endpoint
	teamService := services.NewTeamService(teamRepository, userRepository)
	http.RegisterTeamEndpoints(srv, teamService, []byte(key.Key))

	return userService
}
