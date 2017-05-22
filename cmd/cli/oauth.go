package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/auth/services"
	"github.com/bobinette/papernet/jwt"

	"github.com/bobinette/papernet/oauth"
	"github.com/bobinette/papernet/oauth/bolt"
)

type OAuthConfiguration struct {
	OAuth struct {
		Bolt string `toml:"bolt"`
	} `toml:"oauth"`
}

var (
	// Configuration file
	oauthConfig OAuthConfiguration

	// Other variables
	googleRepository oauth.GoogleRepository
)

func init() {
	OAuthCommand.AddCommand(&OAuthMigrateCommand)

	inheritPersistentPreRun(&OAuthCommand)
	inheritPersistentPreRun(&OAuthMigrateCommand)

	RootCmd.AddCommand(&OAuthCommand)
}

var OAuthCommand = cobra.Command{
	Use:   "oauth",
	Short: "List all the oauth command availables",
	Long:  "List all the oauth command availables",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Read configuration file
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			logger.Fatal("could not read configuration file:", err)
		}

		err = toml.Unmarshal(data, &authConfig)
		if err != nil {
			logger.Fatal("error unmarshalling configuration:", err)
		}

		err = toml.Unmarshal(data, &oauthConfig)
		if err != nil {
			logger.Fatal("error unmarshalling configuration:", err)
		}

		// Read key file
		keyData, err := ioutil.ReadFile(authConfig.Auth.KeyPath)
		if err != nil {
			logger.Fatal("could not open key file:", err)
		}
		// Create token encoder
		var key struct {
			Key string `json:"k"`
		}
		err = json.Unmarshal(keyData, &key)
		if err != nil {
			logger.Fatal("could not read key file:", err)
		}
		tokenEncoder := jwt.NewEncodeDecoder([]byte(key.Key))

		// Create user repository
		store, err := cayley.NewStore(authConfig.Auth.Cayley.Store)
		if err != nil {
			logger.Fatal("could not open user graph:", err)
		}
		userRepository := cayley.NewUserRepository(store)

		// Create user service
		userService = services.NewUserService(userRepository, tokenEncoder)

		// -----
		// Oauth
		driver := &bolt.Driver{}
		if err := driver.Open(oauthConfig.OAuth.Bolt); err != nil {
			logger.Fatal("could not open oauth db:", err)
		}
		googleRepository = bolt.NewGoogleRepository(driver)
	},
}

var OAuthMigrateCommand = cobra.Command{
	Use:   "migrate",
	Short: "Migrate the users in the google oauth database",
	Long:  "Migrate the users in the google oauth database",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := userService.All()
		if err != nil {
			logger.Fatal("error retrieving users:", err)
		}

		for _, user := range users {
			if user.Email == "" || user.GoogleID == "" {
				logger.Errorf("cannot migrate user %d: no email", user.ID)
				continue
			}

			if err := googleRepository.Insert(user.GoogleID, user.ID); err != nil {
				logger.Fatalf("error migrating user %d: %v", user.ID, err)
			}
			logger.Printf("user (%d, %s) migrated", user.ID, user.GoogleID)
		}
	},
}
