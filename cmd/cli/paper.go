package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	ppnBolt "github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/jwt"

	"github.com/bobinette/papernet/auth/cayley"
	authServices "github.com/bobinette/papernet/auth/services"

	"github.com/bobinette/papernet/papernet"
	"github.com/bobinette/papernet/papernet/auth"
	"github.com/bobinette/papernet/papernet/bleve"
	"github.com/bobinette/papernet/papernet/bolt"
	"github.com/bobinette/papernet/papernet/services"
)

type PaperConfig struct {
	Paper struct {
		Bolt struct {
			Store string `toml:"store"`
		} `toml:"bolt"`
		Bleve struct {
			Store string `toml:"store"`
		} `toml:"bleve"`
	} `toml:"paper"`
	// Legacy
	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

var (
	paperConfig PaperConfig

	boltDriver *bolt.Driver

	paperRepository papernet.PaperRepository
	paperIndex      papernet.PaperIndex

	tagService   *services.TagService
	paperService *services.PaperService
)

func init() {

	PaperCommand.AddCommand(&SavePaperCommand)
	PaperCommand.AddCommand(&DeletePaperCommand)
	PaperCommand.AddCommand(&SearchCommand)
	PaperCommand.AddCommand(&PaperMigrateCommand)
	PaperCommand.AddCommand(&PaperFixSequenceCommand)
	PaperCommand.AddCommand(&PaperIndexCommand)
	PaperIndexCommand.AddCommand(&PaperIndexAllCommand)

	inheritPersistentPreRun(&SavePaperCommand)
	inheritPersistentPreRun(&DeletePaperCommand)
	inheritPersistentPreRun(&SearchCommand)
	inheritPersistentPreRun(&PaperMigrateCommand)
	inheritPersistentPreRun(&PaperFixSequenceCommand)
	inheritPersistentPreRun(&PaperIndexCommand)
	inheritPersistentPreRun(&PaperIndexAllCommand)
	inheritPersistentPreRun(&PaperCommand)

	RootCmd.AddCommand(&PaperCommand)
}

var PaperCommand = cobra.Command{
	Use:   "paper",
	Short: "Find papers based on their IDs",
	Long:  "Find papers based on their IDs",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Read configuration file
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			logger.Fatal("could not read configuration file:", err)
		}

		// Load user service
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
		userService = authServices.NewUserService(userRepository, tokenEncoder)

		// Load paper service
		err = toml.Unmarshal(data, &paperConfig)
		if err != nil {
			logger.Fatal("error unmarshalling configuration:", err)
		}

		// Create paper repository and tag index
		boltDriver = &bolt.Driver{}
		if boltDriver.Open(paperConfig.Paper.Bolt.Store); err != nil {
			logger.Fatal("could not open bolt driver:", err)
		}
		paperRepo := bolt.PaperRepository{Driver: boltDriver}
		paperRepository = &paperRepo
		tagIndex := bolt.TagIndex{Driver: boltDriver}

		// Create paper index
		index := &bleve.PaperIndex{}
		if err := index.Open(paperConfig.Paper.Bleve.Store); err != nil {
			logger.Fatal("could not open paper index:", err)
		}
		paperIndex = index

		// Create user client
		authClient := auth.NewClient(userService)

		// Create services
		tagService = services.NewTagService(&tagIndex)
		paperService = services.NewPaperService(paperRepository, paperIndex, authClient, tagService)
	},
}

var DeletePaperCommand = cobra.Command{
	Use:   "delete",
	Short: "Delete papers based on their IDs",
	Long:  "Delete papers based on their IDs",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logger.Fatal("This command expects ids as arguments")
		}

		if args[0] == "help" {
			cmd.Help()
			return
		}

		// @TODO: implement
	},
}

var SavePaperCommand = cobra.Command{
	Use:   "save",
	Short: "Save a paper",
	Long:  "Insert or update a paper based on the argument payload or a file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "help" {
			cmd.Help()
			return
		}

		if len(args) != 1 {
			logger.Fatal("when no filename is specified, the payload must be passed as argument")
		}

		var data []byte
		if strings.HasPrefix(args[0], "@") {
			d, err := ioutil.ReadFile(args[0][1:])
			if err != nil {
				logger.Fatal(err)
			}
			data = d
		} else {
			data = []byte(args[0])
		}

		var paper papernet.Paper
		err := json.Unmarshal(data, &paper)
		if err != nil {
			logger.Fatal("error unmarshalling payload:", err)
		}

		if err := paperRepository.Upsert(&paper); err != nil {
			logger.Errorf("error migrating paper %d: %v", paper.ID, err)
		}

		if err := paperIndex.Index(&paper); err != nil {
			logger.Errorf("error indexing paper %d: %v", paper.ID, err)
		}

		logger.Printf("paper %d inserted", paper.ID)

		cmd.Println(paper)
	},
}

var SearchCommand = cobra.Command{
	Use:   "search",
	Short: "Search papers",
	Long:  "Search papers based on the argument payload or a file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "help" {
			cmd.Help()
			return
		}

		// @TODO: implement
	},
}

var PaperMigrateCommand = cobra.Command{
	Use:   "migrate",
	Short: "Migrate papers to v2",
	Long:  "Migrate papers to v2",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "help" {
			cmd.Help()
			return
		}

		driver := ppnBolt.Driver{}
		defer driver.Close()
		err := driver.Open(paperConfig.Bolt.Store)
		if err != nil {
			logger.Fatal("could not open db:", err)
		}
		paperStore := ppnBolt.PaperStore{Driver: &driver}

		papers, err := paperStore.List()
		if err != nil {
			logger.Fatal("could not get papers:", err)
		}

		for _, paper := range papers {
			paperV2 := papernet.Paper{
				ID:         paper.ID,
				Title:      paper.Title,
				Summary:    paper.Summary,
				Authors:    paper.Authors,
				Tags:       paper.Tags,
				References: paper.References,
				CreatedAt:  paper.CreatedAt,
				UpdatedAt:  paper.UpdatedAt,
			}

			if err := paperRepository.Upsert(&paperV2); err != nil {
				logger.Errorf("error migrating paper %d: %v", paper.ID, err)
				continue
			}

			if err := paperIndex.Index(&paperV2); err != nil {
				logger.Errorf("error indexing paper %d: %v", paper.ID, err)
				continue
			}

			logger.Printf("paper %d migrated", paper.ID)
		}
	},
}

var PaperFixSequenceCommand = cobra.Command{
	Use:   "fix-sequence",
	Short: "Fix id generation for papers v2",
	Long:  "Fix id generation for papers v2",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "help" {
			cmd.Help()
			return
		}

		papers, err := paperRepository.List()
		if err != nil {
			logger.Fatal("error retrieving papers:", err)
		}

		id := 0
		for _, paper := range papers {
			if paper.ID > id {
				id = paper.ID
			}
		}
		if err := boltDriver.ResetSequence(id); err != nil {
			logger.Fatal("error resetting sequence:", err)
		}

		logger.Print("Done, sequence is now:", id)
	},
}

var PaperIndexCommand = cobra.Command{
	Use:   "index",
	Short: "Index a paper",
	Long:  "Index a paper",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "help" {
			cmd.Help()
			return
		}

		ids, err := ints(args)
		if err != nil {
			logger.Fatal("error reading ids:", err)
		}

		papers, err := paperRepository.Get(ids...)
		if err != nil {
			logger.Fatal("error retrieving papers:", err)
		}

		for _, paper := range papers {
			err := paperIndex.Index(&paper)
			if err != nil {
				logger.Errorf("error indexing paper %d: %v", paper.ID, err)
			}

			logger.Printf("indexed paper %d", paper.ID)
		}
	},
}

var PaperIndexAllCommand = cobra.Command{
	Use:   "all",
	Short: "Index all papers",
	Long:  "Index all papers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "help" {
			cmd.Help()
			return
		}

		papers, err := paperRepository.List()
		if err != nil {
			logger.Fatal("error retrieving papers:", err)
		}

		for _, paper := range papers {
			err := paperIndex.Index(&paper)
			if err != nil {
				logger.Errorf("error indexing paper %d: %v", paper.ID, err)
			}

			logger.Printf("indexed paper %d", paper.ID)
		}
	},
}

func ints(strs []string) ([]int, error) {
	ints := make([]int, len(strs))

	for i, str := range strs {
		n, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		ints[i] = n
	}

	return ints, nil
}
