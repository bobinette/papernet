package cmd

import (
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/imports/bolt"
	"github.com/bobinette/papernet/imports/services"
)

type Configuration struct {
	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
}

func Start(srv services.HTTPServer, conf Configuration, logger log.Logger) {
	driver := &bolt.Driver{}
	if err := driver.Open(conf.Bolt.Store); err != nil {
		logger.Fatal("error opening db:", err)
	}
	repo := bolt.NewPaperRepository(driver)

	service := services.NewImportService(repo)
	service.RegisterHTTP(srv)
}
