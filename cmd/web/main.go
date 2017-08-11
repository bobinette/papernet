package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/BurntSushi/toml"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
	"github.com/bobinette/papernet/log"
	"github.com/bobinette/papernet/web"

	"github.com/bobinette/papernet/clients"
	authClient "github.com/bobinette/papernet/clients/auth"

	// packages used for migration to go-kit
	"github.com/bobinette/papernet/jwt"

	kitauth "github.com/bobinette/papernet/auth/cmd"
	kitimports "github.com/bobinette/papernet/imports/cmd"
	kitoauth "github.com/bobinette/papernet/oauth/cmd"
	kitpaper "github.com/bobinette/papernet/papernet/cmd"
)

type Configuration struct {
	Clients struct {
		Auth struct {
			User     string `toml:"user"`
			Password string `toml:"password"`
			BaseURL  string `toml:"baseURL"`
		} `toml:"auth"`
	} `toml:"clients"`

	Auth    kitauth.Configuration    `toml:"auth"`
	Oauth   kitoauth.Configuration   `toml:"oauth"`
	Imports kitimports.Configuration `toml:"imports"`
	Paper   kitpaper.Configuration   `toml:"paper"`

	Bolt struct {
		Store string `toml:"store"`
	} `toml:"bolt"`
	Bleve struct {
		Store string `toml:"store"`
	} `toml:"bleve"`
}

func main() {
	env := flag.String("env", "dev", "environment")
	flag.Parse()

	logger := log.New(*env)

	data, err := ioutil.ReadFile(fmt.Sprintf("configuration/config.%s.toml", *env))
	if err != nil {
		logger.Fatal("could not read configuration file:", err)
	}
	var cfg Configuration
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		logger.Fatal("error unmarshalling configuration:", err)
	}

	// Create repositories
	driver := bolt.Driver{}
	defer driver.Close()
	err = driver.Open(cfg.Bolt.Store)
	if err != nil {
		logger.Fatal("could not open db:", err)
	}

	paperStore := bolt.PaperStore{Driver: &driver}
	tagIndex := bolt.TagIndex{Driver: &driver}

	// Create index
	index := bleve.PaperIndex{}
	err = index.Open(cfg.Bleve.Store)
	defer index.Close()
	if err != nil {
		logger.Fatal("could not open index:", err)
	}

	// Importers
	importer := make(papernet.ImporterRegistry)
	importer.Register("arxiv.org", &papernet.ArxivSpider{})
	importer.Register("medium.com", &papernet.MediumImporter{})

	// Auth
	keyData, err := ioutil.ReadFile(cfg.Auth.KeyPath)
	if err != nil {
		logger.Fatal("could not open key file:", err)
	}
	var key papernet.SigningKey
	err = json.Unmarshal(keyData, &key)
	if err != nil {
		logger.Fatal("could not read key file:", err)
	}

	decoder := jwt.NewEncodeDecoder([]byte(key.Key))
	authenticator := auth.Authenticator{
		Decoder: decoder,
	}

	// Start web server
	addr := ":1705"
	server, err := gin.New(addr, &authenticator)
	if err != nil {
		logger.Fatal("could not start server:", err)
	}

	// *************************************************
	// Migration to go-kit
	// *************************************************

	client := clients.NewClient(
		cfg.Clients.Auth.User,
		cfg.Clients.Auth.Password,
		&http.Client{},
		cfg.Clients.Auth.BaseURL,
	)

	ac := authClient.NewClient(
		client,
		cfg.Clients.Auth.BaseURL,
	)

	// Auth service
	userService := kitauth.Start(server, cfg.Auth, logger)

	// OAuth service
	kitoauth.Start(server, cfg.Oauth, logger, userService)

	// Paper service
	_ = kitpaper.Start(server, cfg.Paper, logger, userService, ac)

	// Imports service
	kitimports.Start(server, cfg.Imports, logger, userService)

	// *************************************************
	// Migration to go-kit
	// *************************************************

	// Oops
	authenticator.Service = userService
	// ----

	// Paper handler
	paperHandler := &web.PaperHandler{
		Store:    &paperStore,
		Index:    &index,
		TagIndex: &tagIndex,

		PaperOwnershipRegistry: userService,
	}
	for _, route := range paperHandler.Routes() {
		server.Register(route)
	}

	// Arxiv handler
	arxivHandler := &web.ArxivHandler{
		Index: &index,
		Store: &paperStore,
	}
	for _, route := range arxivHandler.Routes() {
		server.Register(route)
	}

	// Tag handler
	tagHandler := &web.TagHandler{
		Searcher: &tagIndex,
	}
	for _, route := range tagHandler.Routes() {
		server.Register(route)
	}

	// Import handler
	importHandler := &web.ImportHandler{
		Importer: importer,
	}
	for _, route := range importHandler.Routes() {
		server.Register(route)
	}

	// *************************************************
	// Start server
	// *************************************************

	logger.Print("server started, listening on", addr)
	logger.Fatal(server.Start())
}
