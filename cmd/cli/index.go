package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/spf13/cobra"

	"github.com/bobinette/papernet/errors"
)

func init() {
	CreateIndexCommand.PersistentFlags().String("file", "", "mapping file")
	CreateIndexCommand.PersistentFlags().String("index", "", "index directory")

	IndexCommand.AddCommand(&CreateIndexCommand)
	IndexCommand.AddCommand(&IndexAllCommand)
	RootCmd.AddCommand(&IndexCommand)
}

var IndexCommand = cobra.Command{
	Use:   "index",
	Short: "Index papers",
	Long:  "Index papers from their IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("This command expects ids as arguments")
		}

		if args[0] == "help" {
			return cmd.Help()
		}

		ids, err := ints(args)
		if err != nil {
			return errors.New("ids should be integers", errors.WithCause(err))
		}

		addr := cmd.Flag("store").Value.String()
		store, f, err := createStore(addr)
		defer f()
		if err != nil {
			return errors.New("error opening db", errors.WithCause(err))
		}

		addr = cmd.Flag("index").Value.String()
		index, f, err := createIndex(addr)
		defer f()
		if err != nil {
			return errors.New("error opening index", errors.WithCause(err))
		}

		papers, err := store.Get(ids...)
		if err != nil {
			return errors.New("error getting papers", errors.WithCause(err))
		}

		for _, paper := range papers {
			err = index.Index(paper)
			if err != nil {
				return errors.New("error indexing", errors.WithCause(err))
			}
			cmd.Printf("<Paper %d> indexed\n", paper.ID)
		}
		return nil
	},
}

var IndexAllCommand = cobra.Command{
	Use:   "all",
	Short: "Index all papers",
	Long:  "Index all papers in the store",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "help" {
			return cmd.Help()
		}

		addr := cmd.Flag("store").Value.String()
		store, f, err := createStore(addr)
		defer f()
		if err != nil {
			return errors.New("error opening db", errors.WithCause(err))
		}

		addr = cmd.Flag("index").Value.String()
		index, f, err := createIndex(addr)
		defer f()
		if err != nil {
			return errors.New("error opening index", errors.WithCause(err))
		}

		papers, err := store.List()
		if err != nil {
			return errors.New("error getting papers", errors.WithCause(err))
		}

		for _, paper := range papers {
			err = index.Index(paper)
			if err != nil {
				return errors.New("error indexing", errors.WithCause(err))
			}
			cmd.Printf("<Paper %d> indexed\n", paper.ID)
		}
		return nil
	},
}

var CreateIndexCommand = cobra.Command{
	Use:   "create",
	Short: "Create index",
	Long:  "Create index",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "help" {
			return cmd.Help()
		}

		filename := cmd.Flag("mapping").Value.String()
		if filename == "" {
			return errors.New("mapping argument needed")
		}

		indexPath := cmd.Flag("index").Value.String()
		if indexPath == "" {
			return errors.New("index argument needed")
		}

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.New("error reading mapping file", errors.WithCause(err))
		}

		var m mapping.IndexMappingImpl
		err = json.Unmarshal(data, &m)

		_, err = bleve.New(indexPath, &m)
		if err != nil {
			return errors.New("error creating index", errors.WithCause(err))
		}
		cmd.Printf("Index created at %s\n", indexPath)

		return nil
	},
}
