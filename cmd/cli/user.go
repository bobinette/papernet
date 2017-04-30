package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/errors"
)

func init() {
	UserCommand.PersistentFlags().String("store", "data/papernet.db", "address of the bolt db file")

	UserCommand.AddCommand(&CreateUserCommand)
	UserCommand.AddCommand(&UserAllCommand)
	RootCmd.AddCommand(&UserCommand)
}

var UserCommand = cobra.Command{
	Use:   "user",
	Short: "List all users",
	Long:  "List all users",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatalln("not ready")
	},
}

var CreateUserCommand = cobra.Command{
	Use:   "create",
	Short: "Create a user",
	Long:  "Create a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("takes only one argument: the user in json")
		}
		if args[0] == "help" {
			return cmd.Help()
		}

		addr := cmd.Flag("store").Value.String()
		driver := bolt.Driver{}

		err := driver.Open(addr)
		defer driver.Close()
		if err != nil {
			return errors.New("error opening db", errors.WithCause(err))
		}

		var user papernet.User
		err = json.NewDecoder(strings.NewReader(args[0])).Decode(&user)
		if err != nil {
			return err
		}

		store := bolt.UserStore{Driver: &driver}
		err = store.Upsert(&user)
		if err != nil {
			return err
		}

		log.Println("done")
		return nil
	},
}

var UserAllCommand = cobra.Command{
	Use:   "all",
	Short: "List all users",
	Long:  "List all users",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] == "help" {
			cmd.Help()
			return
		}

		addr := cmd.Flag("store").Value.String()
		driver := bolt.Driver{}

		err := driver.Open(addr)
		defer driver.Close()
		if err != nil {
			log.Fatalln(errors.New("error opening db", errors.WithCause(err)))
		}

		store := bolt.UserStore{Driver: &driver}

		users, err := store.List()
		if err != nil {
			log.Fatalln(errors.New("error getting papers", errors.WithCause(err)))
		}
		for _, user := range users {
			data, err := json.Marshal(user)
			if err != nil {
				log.Fatalln(err)
			}
			cmd.Println(string(data))
		}
	},
}
