package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/errors"
)

func init() {
	TeamCommand.PersistentFlags().String("store", "data/papernet.db", "address of the bolt db file")
	RootCmd.AddCommand(&TeamCommand)
}

var TeamCommand = cobra.Command{
	Use:   "team",
	Short: "List all teams",
	Long:  "List all teams",
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

		store := bolt.TeamStore{Driver: &driver}

		teams, err := store.All()
		if err != nil {
			return errors.New("error getting teams", errors.WithCause(err))
		}
		for _, team := range teams {
			data, err := json.Marshal(team)
			if err != nil {
				log.Println(err)
				fmt.Println(team)
				continue
			}

			fmt.Println(string(data))
		}

		return nil
	},
}
