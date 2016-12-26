package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
)

func createStore(addr string) (papernet.PaperStore, func(), error) {
	driver := bolt.Driver{}

	err := driver.Open(addr)
	if err != nil {
		return nil, func() {}, err
	}

	store := bolt.PaperStore{Driver: &driver}
	return &store, func() { driver.Close() }, nil
}

func createIndex(addr string) (papernet.PaperIndex, func(), error) {
	index := bleve.PaperIndex{}
	err := index.Open(addr)
	if err != nil {
		return nil, func() {}, err
	}

	return &index, func() { index.Close() }, nil
}

func reindexAll(store papernet.PaperStore, index papernet.PaperIndex, v bool) error {
	papers, err := store.List()
	if err != nil {
		return err
	}

	for _, paper := range papers {
		err := index.Index(paper)
		if err != nil {
			return err
		}

		if v {
			log.Println("Paper", paper.ID, "reindexed")
		}
	}

	return nil
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
