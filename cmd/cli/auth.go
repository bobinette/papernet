package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/auth/services"
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
	userRepository auth.UserRepository
	userService    *services.UserService
	teamService    *services.TeamService
)

func init() {
	AuthCommand.AddCommand(&AuthUserCommand)

	AuthUserCommand.AddCommand(&AuthTokenCommand)
	AuthUserCommand.AddCommand(&AuthAllUsersCommand)
	AuthUserCommand.AddCommand(&AuthUpsertUserCommand)
	AuthUserCommand.AddCommand(&AuthDeleteCommand)

	inheritPersistentPreRun(&AuthCommand)
	inheritPersistentPreRun(&AuthUserCommand)
	inheritPersistentPreRun(&AuthTokenCommand)
	inheritPersistentPreRun(&AuthUpsertUserCommand)
	inheritPersistentPreRun(&AuthDeleteCommand)

	inheritPersistentPreRun(&AuthAllUsersCommand)

	RootCmd.AddCommand(&AuthCommand)
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
		tokenEncoder := jwt.NewEncodeDecoder([]byte(key.Key))

		// Create user repository
		store, err := cayley.NewStore(authConfig.Auth.Cayley.Store)
		if err != nil {
			logger.Fatal("could not open user graph:", err)
		}
		userRepository = cayley.NewUserRepository(store)
		teamRepository := cayley.NewTeamRepository(store)

		// Create user service
		userService = services.NewUserService(userRepository, tokenEncoder)
		teamService = services.NewTeamService(teamRepository, userRepository)
	},
}

var AuthUserCommand = cobra.Command{
	Use:   "user",
	Short: "Retrieve a user based on its id",
	Long:  "Retrieve a user based on its id",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatal("user 1 argument: the id of the user")
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

var AuthDeleteCommand = cobra.Command{
	Use: "delete",
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

		err = userService.Delete(userID)
		if err != nil {
			logger.Fatal(err)
		}
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

var AuthUpsertUserCommand = cobra.Command{
	Use:   "upsert",
	Short: "Upsert a user from a json",
	Long:  "Upsert a user from a json",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatalf("user upsert wants 1 argument: the json to upsert the user, got %v", args)
		}

		if args[0] == "help" {
			cmd.Help()
			return
		}

		var user auth.User
		err := json.Unmarshal([]byte(args[0]), &user)
		if err != nil {
			logger.Fatal(err)
		}

		user, err = userService.Upsert(user)
		if err != nil {
			logger.Fatal(err)
		}

		data, err := json.Marshal(user)
		if err != nil {
			logger.Fatal(err)
		}

		cmd.Println("user upserted: ", string(data))
	},
}
