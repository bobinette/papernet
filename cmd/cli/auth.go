package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/auth/cayley"
	"github.com/bobinette/papernet/auth/services"
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
	userService *services.UserService
	teamService *services.TeamService
)

func init() {
	AuthCommand.AddCommand(&AuthUserCommand)

	AuthUserCommand.AddCommand(&AuthTokenCommand)
	AuthUserCommand.AddCommand(&AuthAllUsersCommand)
	AuthUserCommand.AddCommand(&AuthUpsertUserCommand)

	AuthCommand.AddCommand(&AuthMigrationCommand)

	inheritPersistentPreRun(&AuthCommand)
	inheritPersistentPreRun(&AuthUserCommand)
	inheritPersistentPreRun(&AuthTokenCommand)
	inheritPersistentPreRun(&AuthMigrationCommand)
	inheritPersistentPreRun(&AuthUpsertUserCommand)

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
		userRepository := cayley.NewUserRepository(store)
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

		// ------------------------------------------------
		// Migrate Users
		userStore := bolt.UserStore{Driver: &driver}
		paperStore := bolt.PaperStore{Driver: &driver}
		teamStore := bolt.TeamStore{Driver: &driver}

		oldUsers, err := userStore.List()
		if err != nil {
			logger.Fatal(errors.New("error getting papers", errors.WithCause(err)))
		}

		users := make(map[int]auth.User)
		usersByGoogleID := make(map[string]auth.User)
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
			users[user.ID] = user
			usersByGoogleID[user.GoogleID] = user

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

		paperOwner := make(map[int]int)
		for paperID, userIDs := range paperPermission {
			papers, err := paperStore.Get(paperID)
			if err != nil {
				logger.Fatal(err)
			} else if len(papers) == 0 {
				cmd.Printf("paper %d does not exist anymore.\n", paperID)
				continue
			}

			if len(userIDs) == 1 {
				_, err := userService.CreatePaper(userIDs[0], paperID)
				if err != nil {
					logger.Fatal(err)
				}
				paperOwner[paperID] = userIDs[0]
				cmd.Printf("paper %d attributed to user %d\n", paperID, userIDs[0])
			} else {
				paper := papers[0]
				cmd.Printf("Conflict of owner for paper %d: %s\n", paper.ID, paper.Title)
				for _, userID := range userIDs {
					cmd.Printf("%d: %s\n", users[userID].ID, users[userID].Name)
				}

				cmd.Print("Owner: ")
				var userID int
				fmt.Scanln(&userID)

				valid := false
				for _, uID := range userIDs {
					if uID == userID {
						valid = true
						break
					}
				}
				if !valid {
					logger.Fatalf("invalid user id %d, not in %v", userID, userIDs)
				}

				_, err = userService.CreatePaper(userID, paperID)
				if err != nil {
					logger.Fatal(err)
				}
				cmd.Printf("paper %d attributed to user %d\n", paperID, userID)
				paperOwner[paperID] = userID
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

		// ------------------------------------------------
		// Migrate Teams

		oldTeams, err := teamStore.All()
		if err != nil {
			logger.Fatal(err)
		}

		for _, oldTeam := range oldTeams {
			team := auth.Team{
				Name: oldTeam.Name,
			}
			adminGoogleID := oldTeam.Admins[0]
			admin := usersByGoogleID[adminGoogleID]

			team, err := teamService.Create(admin.ID, team)
			if err != nil {
				logger.Fatal(err)
			}

			for _, memberGoogleID := range oldTeam.Members {
				team, err = teamService.Invite(admin.ID, team.ID, usersByGoogleID[memberGoogleID].Email)
				if err != nil {
					logger.Fatal(err)
				}
			}

			for _, paperID := range oldTeam.CanSee {
				canEdit := false
				for _, pID := range oldTeam.CanEdit {
					if pID == paperID {
						canEdit = true
						break
					}
				}

				owner, ok := paperOwner[paperID]
				if !ok {
					cmd.Printf("no owner found for paper %d, paper has probably been deleted\n", paperID)
					continue
				}

				team, err = teamService.Share(owner, team.ID, paperID, canEdit)
				if err != nil {
					logger.Fatal(err)
				}
			}

			data, err := json.Marshal(team)
			if err != nil {
				logger.Fatal(err)
			}
			cmd.Printf("team %d created: %s\n", team.ID, data)
		}
	},
}
