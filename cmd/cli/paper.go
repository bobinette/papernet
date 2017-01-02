package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/errors"
)

func init() {
	SavePaperCommand.PersistentFlags().String("file", "", "filename to load the payload")
	SearchCommand.PersistentFlags().String("file", "", "filename to load the payload")

	PaperCommand.AddCommand(&SavePaperCommand)
	PaperCommand.AddCommand(&DeletePaperCommand)
	PaperCommand.AddCommand(&SearchCommand)

	RootCmd.AddCommand(&PaperCommand)
}

var PaperCommand = cobra.Command{
	Use:   "paper",
	Short: "Find papers based on their IDs",
	Long:  "Find papers based on their IDs",
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

		papers, err := store.Get(ids...)
		if err != nil {
			return errors.New("error getting papers", errors.WithCause(err))
		}

		pj, err := json.Marshal(papers)
		if err != nil {
			return errors.New("error marshalling results", errors.WithCause(err))
		}

		cmd.Println(string(pj))
		return nil
	},
}

var DeletePaperCommand = cobra.Command{
	Use:   "delete",
	Short: "Delete papers based on their IDs",
	Long:  "Delete papers based on their IDs",
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

		for _, id := range ids {
			err = store.Delete(id)
			if err != nil {
				return errors.New("error deleting in store papers", errors.WithCause(err))
			}

			err = index.Delete(id)
			if err != nil {
				return errors.New("error deleting in index papers", errors.WithCause(err))
			}

			cmd.Printf("<Paper %d> deleted\n", id)
		}
		return nil
	},
}

var SavePaperCommand = cobra.Command{
	Use:   "save",
	Short: "Save a paper",
	Long:  "Insert or update a paper based on the argument payload or a file",
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

		filename := cmd.Flag("file").Value.String()
		var data []byte
		if filename != "" {
			data, err = ioutil.ReadFile(filename)
			if err != nil {
				return errors.New("error reading payload file", errors.WithCause(err))
			}
		} else {
			if len(args) != 1 {
				return errors.New("when no filename is specified, the payload must be passed as argument")
			}
			data = []byte(args[0])
		}

		var paper papernet.Paper
		err = json.Unmarshal(data, &paper)
		if err != nil {
			return errors.New("error unmarshalling payload", errors.WithCause(err))
		}

		err = store.Upsert(&paper)
		if err != nil {
			return errors.New("error saving paper", errors.WithCause(err))
		}

		cmd.Println("done")
		return nil
	},
}

var SearchCommand = cobra.Command{
	Use:   "search",
	Short: "Search papers",
	Long:  "Search papers based on the argument payload or a file",
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

		filename := cmd.Flag("file").Value.String()
		var data []byte
		if filename != "" {
			data, err = ioutil.ReadFile(filename)
			if err != nil {
				return errors.New("error reading payload file", errors.WithCause(err))
			}
		} else {
			if len(args) != 1 {
				return errors.New("when no filename is specified, the payload must be passed as argument")
			}
			data = []byte(args[0])
		}

		var search papernet.PaperSearch
		err = json.Unmarshal(data, &search)
		if err != nil {
			return errors.New("error unmarshalling payload", errors.WithCause(err))
		}

		res, err := index.Search(search)
		if err != nil {
			return errors.New("error querying index", errors.WithCause(err))
		}

		papers, err := store.Get(res.IDs...)
		if err != nil {
			return errors.New("error retrieving papers", errors.WithCause(err))
		}

		pj, err := json.Marshal(papers)
		if err != nil {
			return errors.New("error marshalling results", errors.WithCause(err))
		}

		cmd.Println(string(pj))
		return nil
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
