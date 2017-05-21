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
