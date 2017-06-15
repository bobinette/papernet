package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet"
	ppnBolt "github.com/bobinette/papernet/bolt"

	"github.com/bobinette/papernet/imports"
	"github.com/bobinette/papernet/imports/bolt"
)

type ImportsConfiguration struct {
	Imports struct {
		Bolt struct {
			Store string `toml:"store"`
		} `toml:"bolt"`
	} `toml:"imports"`
	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

var (
	importsConfiguration ImportsConfiguration

	paperStore        papernet.PaperStore
	importsRepository imports.Repository
)

func init() {
	ImportsCommand.AddCommand(&ImportsMigrateCommand)

	inheritPersistentPreRun(&ImportsCommand)
	inheritPersistentPreRun(&ImportsMigrateCommand)

	RootCmd.AddCommand(&ImportsCommand)
}

var ImportsCommand = cobra.Command{
	Use:   "imports",
	Short: "List all the imports command availables",
	Long:  "List all the imports command availales",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load auth service
		AuthCommand.PersistentPreRun(cmd, args)

		// Read configuration file
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			logger.Fatal("could not read configuration file:", err)
		}

		err = toml.Unmarshal(data, &importsConfiguration)
		if err != nil {
			logger.Fatal("error unmarshalling configuration:", err)
		}

		ppndDriver := ppnBolt.Driver{}
		err = ppndDriver.Open(importsConfiguration.Bolt.Store)
		if err != nil {
			logger.Fatal("could not open paper driver:", err)
		}
		paperStore = &ppnBolt.PaperStore{Driver: &ppndDriver}

		driver := bolt.Driver{}
		err = driver.Open(importsConfiguration.Imports.Bolt.Store)
		if err != nil {
			logger.Fatal("could not open imports driver:", err)
		}
		importsRepository = bolt.NewPaperRepository(&driver)
	},
}

var ImportsMigrateCommand = cobra.Command{
	Use: "migrate",
	Run: func(cmd *cobra.Command, args []string) {
		papers, err := paperStore.List()
		if err != nil {
			logger.Fatal("could not read papers:", err)
		}

		for _, paper := range papers {
			if paper.ArxivID == "" {
				continue
			}

			userID, err := userRepository.PaperOwner(paper.ID)
			if err != nil {
				logger.Printf("could not migrate paper %d:", err)
				continue
			} else if userID == 0 {
				logger.Printf("could not migrate paper %d: owner not found")
				continue

			}
			importsRepository.Save(userID, paper.ID, "arxiv", paper.ArxivID)
			logger.Printf("Paper %d migrated (arxiv ID: %s", paper.ID, paper.ArxivID)
		}
	},
}
