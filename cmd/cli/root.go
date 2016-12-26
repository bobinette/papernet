package main

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.PersistentFlags().String("store", "data/papernet.db", "address of the bolt db file")
	RootCmd.PersistentFlags().String("index", "data/papernet.index", "address of the bolt db file")
}

var RootCmd = cobra.Command{
	Use:   "papernet",
	Short: "Keep track of your knowledge with Papernet",
	Long:  "Keep track of your knowledge with Papernet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
