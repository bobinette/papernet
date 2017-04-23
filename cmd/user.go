package main

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/auth"
)

func init() {
	UserCommand.AddCommand(&UserAllCommand)
	UserCommand.AddCommand(&UserUpsertCommand)
	UserCommand.AddCommand(&TokenCommand)
	RootCmd.AddCommand(&UserCommand)
}

var UserCommand = cobra.Command{
	Use:   "user",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatal("user wants 1 argument: the id of the user")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Fatal("error converting user id: ", err)
		}

		user, err := userService.Get(id)
		if err != nil {
			logger.Fatal("error retrieving user:", err)
		}

		data, err := formatUser(user)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Print(data)
	},
}

var UserAllCommand = cobra.Command{
	Use:   "all",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := userService.List()
		if err != nil {
			logger.Fatal("error listing users:", err)
		}

		for _, user := range users {
			data, err := formatUser(user)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Print(data)
		}
	},
}

var UserUpsertCommand = cobra.Command{
	Use:   "upsert",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatal("user upsert wants 1 argument: the json representation of the user")
		}

		var u auth.User
		if err := json.Unmarshal([]byte(args[0]), &u); err != nil {
			logger.Fatal("error decoding ruesteq:", err)
		}

		user, err := userService.Upsert(u)
		if err != nil {
			logger.Fatal("error upserting user:", err)
		}

		data, err := formatUser(user)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Print(data)
	},
}

var TokenCommand = cobra.Command{
	Use:   "token",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Fatal("user token wants 1 argument: the id of the user")
		}

		if args[0] == "help" {
			cmd.Help()
			return
		}

		userID := args[0]
		token, err := tokenEncoder.Encode(userID)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Print(token)
	},
}

func formatUser(user auth.User) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
