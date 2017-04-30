package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/jwt"
)

type AuthConfiguration struct {
	Auth struct {
		KeyPath string `toml:"key"`
		Cayley  struct {
			Store string `toml:"store"`
		} `toml:"cayley"`
	} `toml:"auth"`
	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

var (
	// Configuration file
	authConfig AuthConfiguration

	// Other variables
	userService *auth.UserService
)

func init() {
	AuthCommand.AddCommand(&AuthUserCommand)
	AuthCommand.AddCommand(&AuthTokenCommand)
	AuthCommand.AddCommand(&AuthMigrationCommand)

	AuthUserCommand.AddCommand(&AuthAllUsersCommand)

	inheritPersistentPreRun(&AuthCommand)
	inheritPersistentPreRun(&AuthUserCommand)
	inheritPersistentPreRun(&AuthTokenCommand)
	inheritPersistentPreRun(&AuthMigrationCommand)

	inheritPersistentPreRun(&AuthAllUsersCommand)

	RootCmd.AddCommand(&AuthCommand)
}

func inheritPersistentPreRun(cmd *cobra.Command) {
	ppr := cmd.PersistentPreRun
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		// Run parent persistent pre run
		if cmd.Parent() != nil && cmd.Parent().PersistentPreRun != nil {
			cmd.Parent().PersistentPreRun(c, args)
		}

		// Run command persistent pre run
		if ppr != nil {
			ppr(c, args)
		}
	}
}

var AuthCommand = cobra.Command{
	Use:   "auth",
	Short: "List all the auth command availables",
	Long:  "List all the auth command availales",
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
		tokenEncoder := jwt.NewEncoder([]byte(key.Key))

		// Create user repository
		store, err := cayley.NewStore(authConfig.Auth.Cayley.Store)
		if err != nil {
			logger.Fatal("could not open user graph:", err)
		}
		userRepository := cayley.NewUserRepository(store)

		// Create user service
		userService = auth.NewUserService(userRepository, tokenEncoder)
	},
}

var AuthUserCommand = cobra.Command{
	Use:   "user",
	Short: "Retrieve a user based on its id",
	Long:  "Retrieve a user based on its id",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatal("user token wants 1 argument: the id of the user")
		}

		if args[0] == "help" {
			cmd.Help()
			return
		}

		userID, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		user, err := userService.Get(userID)
		if err != nil {
			logger.Fatal(err)
		}

		data, err := json.Marshal(user)
		if err != nil {
			logger.Fatal(err)
		}
		cmd.Println(string(data))
	},
}

var AuthTokenCommand = cobra.Command{
	Use:   "token",
	Short: "Craft a token for a user",
	Long:  "Craft a token for a user",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatal("user token wants 1 argument: the id of the user")
		}

		if args[0] == "help" {
			cmd.Help()
			return
		}

		userID, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		token, err := userService.Token(userID)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Print(token)
	},
}

var AuthAllUsersCommand = cobra.Command{
	Use:   "all",
	Short: "Retrieve all the users",
	Long:  "Retrieve all the users",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] == "help" {
			cmd.Help()
			return
		}

		users, err := userService.All()
		if err != nil {
			logger.Fatal(err)
		}

		data, err := json.Marshal(users)
		if err != nil {
			logger.Fatal(err)
		}
		cmd.Println(string(data))
	},
}

// ------------------------------------------------------------------------------------
// Migration command

var AuthMigrationCommand = cobra.Command{
	Use:   "migrate",
	Short: "Handle the migration to auth v2",
	Long:  "Handle the migration to auth v2 by copying the users and teams",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] == "help" {
			cmd.Help()
			return
		}
		driver := bolt.Driver{}
		err := driver.Open(authConfig.Bolt.Store)
		defer driver.Close()
		if err != nil {
			logger.Fatal(errors.New("error opening db", errors.WithCause(err)))
		}
		userBoltStore := bolt.UserStore{Driver: &driver}

		oldUsers, err := userBoltStore.List()
		if err != nil {
			logger.Fatal(errors.New("error getting papers", errors.WithCause(err)))
		}

		paperPermission := make(map[int][]int)
		userBookmarks := make(map[int][]int)

		for _, oldUser := range oldUsers {
			user := auth.User{
				Name:      oldUser.Name,
				Email:     oldUser.Email,
				GoogleID:  oldUser.ID,
				Owns:      oldUser.CanSee,
				Bookmarks: oldUser.Bookmarks,
			}
			user, err = userService.Upsert(user)
			if err != nil {
				logger.Fatal(err)
			}

			for _, paperID := range oldUser.CanEdit {
				paperPermission[paperID] = append(paperPermission[paperID], user.ID)
			}

			userBookmarks[user.ID] = oldUser.Bookmarks

			data, err := json.Marshal(user)
			if err != nil {
				logger.Fatal(err)
			}
			cmd.Printf("user %d created: %s\n", user.ID, data)
		}

		for paperID, userIDs := range paperPermission {
			if len(userIDs) == 1 {
				_, err := userService.CreatePaper(userIDs[0], paperID)
				if err != nil {
					logger.Fatal(err)
				}
				cmd.Printf("paper %d attributed to user %d\n", paperID, userIDs[0])
			} else {
				cmd.Printf("conflict of owner on paper %d between users %v\n", paperID, userIDs)
			}
		}

		for userID, bookmarks := range userBookmarks {
			for _, paperID := range bookmarks {
				_, err := userService.Bookmark(userID, paperID, true)
				if err != nil {
					cmd.Printf("could not bookmark paper %d for user %d: %v\n", paperID, userID, err)
				} else {
					cmd.Printf("paper %d bookmarked for user %d\n", paperID, userID)
				}
			}
		}
	},
}
