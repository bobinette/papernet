package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
	"github.com/bobinette/papernet/log"
	"github.com/bobinette/papernet/web"

	// packages used for migration to go-kit
	kitauth "github.com/bobinette/papernet/auth/cmd"
	"github.com/bobinette/papernet/jwt"
	"github.com/bobinette/papernet/oauth"
)

type Configuration struct {
	Auth  kitauth.Configuration `toml:"auth"`
	Oauth oauth.Configuration   `toml:"oauth"`
	Bolt  struct {
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

	// Auth service
	userService := kitauth.Start(server, cfg.Auth, logger)

	// OAuth service
	authUserService := oauth.NewUserClient(userService)
	googleService, err := oauth.NewGoogleService(cfg.Oauth.GooglePath, authUserService)
	if err != nil {
		logger.Fatal("could not instantiate google service")
	}
	oauth.RegisterHTTPRoutes(server, googleService)

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
