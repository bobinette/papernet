package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/log"
)

var (
	// flags
	verbose bool
	env     string

	// logger
	logger log.Logger

	// auth
	tokenEncoder auth.TokenEncoder

	// drivers
	boltDriver *bolt.Driver

	// stores
	paperStore papernet.PaperStore
	userStore  papernet.UserStore
	teamStore  papernet.TeamStore

	// indices
	paperIndex papernet.PaperIndex

	// services
	userService *auth.UserService
)

type Configuration struct {
	Auth struct {
		Key    string `toml:"key"`
		Google string `toml:"google"`
	} `toml:"auth"`
	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
	Bleve struct {
		Store string `toml:"store"`
	} `toml:"bleve"`
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	RootCmd.PersistentFlags().StringVar(&env, "env", "dev", "")
}

var RootCmd = cobra.Command{
	Use:   "papernet",
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfgData, err := ioutil.ReadFile(fmt.Sprintf("configuration/config.%s.toml", env))
		if err != nil {
			fmt.Println("error reading configuration:", err)
			return
		}

		var cfg Configuration
		err = toml.Unmarshal(cfgData, &cfg)
		if err != nil {
			fmt.Println("error unmarshalling configuration:", err)
			return
		}

		// Create logger
		logger = log.New(env)

		// Create encoder
		keyData, err := ioutil.ReadFile(cfg.Auth.Key)
		if err != nil {
			logger.Fatal("could not open key file:", err)
		}
		var key papernet.SigningKey
		err = json.Unmarshal(keyData, &key)
		if err != nil {
			logger.Fatal("could not read key file:", err)
		}
		tokenEncoder = auth.EncodeDecoder{Key: key.Key}

		// Create stores
		boltDriver = &bolt.Driver{}
		boltDriver.Open(cfg.Bolt.Store)
		paperStore = &bolt.PaperStore{Driver: boltDriver}
		userStore = &bolt.UserStore{Driver: boltDriver}
		teamStore = &bolt.TeamStore{Driver: boltDriver}

		// Create services
		// -- user service
		userRepository, err := cayley.New("data/user.graph")
		if err != nil {
			logger.Fatal("could not create user graph:", err)
		}
		userService = auth.NewUserService(userRepository)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		boltDriver.Close()
	},
}
