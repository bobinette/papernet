package cmd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/imports/arxiv"
	"github.com/bobinette/papernet/imports/bolt"
	"github.com/bobinette/papernet/imports/http"
	"github.com/bobinette/papernet/imports/services"
)

type Configuration struct {
	KeyPath  string `toml:"key"`
	PaperURL string `toml:"paperURL"`
	Bolt     struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

func Start(srv services.HTTPServer, conf Configuration, logger log.Logger, us http.UserService) {
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
	paperService := http.NewPaperService(us, conf.PaperURL)

	// Searchers
	// Arxiv
	arxivSearcher := arxiv.NewSearcher()

	service := services.NewImportService(repo, paperService, arxivSearcher)
	service.RegisterHTTP(srv, []byte(key.Key))
}
