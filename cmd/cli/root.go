package main

import (
	"github.com/spf13/cobra"
)

func init() {
}

var RootCmd = cobra.Command{
	Use:   "papernet",
	Short: "Keep track of your knowledge with Papernet",
	Long:  "Keep track of your knowledge with Papernet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
