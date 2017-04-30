package main

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/log"
)

var (
	// flags
	env        string
	configFile string

	// logger
	logger log.Logger
)

func init() {
	RootCmd.PersistentFlags().StringVar(&env, "env", "dev", "environment")
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "configuration file")
}

var RootCmd = cobra.Command{
	Use:   "papernet",
	Short: "Keep track of your knowledge with Papernet",
	Long:  "Keep track of your knowledge with Papernet",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger = log.New(env)

		if configFile == "" {
			configFile = path.Join("configuration", fmt.Sprintf("config.%s.toml", env))
		}
	},
}
