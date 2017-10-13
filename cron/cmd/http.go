package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/clients/imports"
	"github.com/bobinette/papernet/log"

	"github.com/bobinette/papernet/cron"
	"github.com/bobinette/papernet/cron/mail"
	"github.com/bobinette/papernet/cron/mysql"
)

type Configuration struct {
	KeyPath string `toml:"key"`
	MySQL   struct {
		Host     string `toml:"host"`
		Port     string `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
		Database string `toml:"database"`
	} `toml:"mysql"`
}

func Start(
	srv cron.HTTPServer,
	conf Configuration,
	logger log.Logger,
	authClient *auth.Client,
	imporstClient *imports.Client,
) {
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

	driver, err := mysql.NewDriver(
		conf.MySQL.Host,
		conf.MySQL.Port,
		conf.MySQL.User,
		conf.MySQL.Password,
		conf.MySQL.Database,
	)
	if err != nil {
		logger.Fatal("error connecting to MySQL:", err)
	}

	repo := mysql.NewRepository(driver)
	resultsRepo := mysql.NewResultsRepository(driver)
	notifierFactory := mail.NewNotifierFactory(authClient)

	service := cron.NewService(repo, resultsRepo, notifierFactory, imporstClient, logger)
	service.RegisterHTTP(srv, []byte(key.Key), authClient)

	service.StartCron(context.Background())
}
