package cmd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/clients/paper"
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/imports"
	"github.com/bobinette/papernet/imports/arxiv"
	"github.com/bobinette/papernet/imports/bolt"
)

type Configuration struct {
	KeyPath string `toml:"key"`
	Bolt    struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

func Start(srv imports.HTTPServer, conf Configuration, logger log.Logger, paperClient *paper.Client, authClient *auth.Client) {
	// Load key from file
	keyData, err := ioutil.ReadFile(conf.KeyPath)
	if err != nil {
		logger.Fatal("could not open key file:", err)
	}

	// Extract key from data
	var key struct {
		Key string `json:"k"`
	}
	err = json.Unmarshal(keyData, &key)
	if err != nil {
		logger.Fatal("could not read key file:", err)
	}

	driver := &bolt.Driver{}
	if err := driver.Open(conf.Bolt.Store); err != nil {
		logger.Fatal("error opening db:", err)
	}
	repo := bolt.NewPaperRepository(driver)

	// Searchers
	// Arxiv
	arxivSearcher := arxiv.NewSearcher()

	service := imports.NewService(repo, paperClient, arxivSearcher)
	service.RegisterHTTP(srv, []byte(key.Key), authClient)
}
