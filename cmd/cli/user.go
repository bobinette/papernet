package main

import (
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/errors"
)

func init() {
	UserCommand.PersistentFlags().String("store", "data/papernet.db", "address of the bolt db file")

	RootCmd.AddCommand(&UserCommand)
}

var UserCommand = cobra.Command{
	Use:   "user",
	Short: "List all users",
	Long:  "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && args[0] == "help" {
			return cmd.Help()
		}

		addr := cmd.Flag("store").Value.String()
		driver := bolt.Driver{}

		err := driver.Open(addr)
		defer driver.Close()
		if err != nil {
			return errors.New("error opening db", errors.WithCause(err))
		}

		store := bolt.UserRepository{Driver: &driver}

		users, err := store.List()
		if err != nil {
			return errors.New("error getting papers", errors.WithCause(err))
		}
		for _, user := range users {
			cmd.Printf("%+v\n", user)
		}

		return nil
	},
}
