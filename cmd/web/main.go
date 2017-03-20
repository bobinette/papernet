package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
	"github.com/bobinette/papernet/web"
)

type Configuration struct {
	Auth struct {
		Key    string `toml:"key"`
		Google string `toml:"google"`
	} `toml:"auth"`
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

	data, err := ioutil.ReadFile(fmt.Sprintf("configuration/config.%s.toml", *env))
	if err != nil {
		log.Fatalln("could not read configuration file:", err)
	}
	var cfg Configuration
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalln("error unmarshalling configuration:", err)
	}

	// Create repositories
	driver := bolt.Driver{}
	defer driver.Close()
	err = driver.Open(cfg.Bolt.Store)
	if err != nil {
		log.Fatalln("could not open db:", err)
	}

	paperStore := bolt.PaperStore{Driver: &driver}
	tagIndex := bolt.TagIndex{Driver: &driver}
	userStore := bolt.UserStore{Driver: &driver}
	teamStore := bolt.TeamStore{Driver: &driver}

	// Create index
	index := bleve.PaperIndex{}
	err = index.Open(cfg.Bleve.Store)
	defer index.Close()
	if err != nil {
		log.Fatalln("could not open index:", err)
	}

	// Importers
	importer := make(papernet.ImporterRegistry)
	importer.Register("arxiv.org", &papernet.ArxivSpider{})
	importer.Register("medium.com", &papernet.MediumImporter{})

	// Auth
	keyData, err := ioutil.ReadFile(cfg.Auth.Key)
	if err != nil {
		log.Fatalln("could not open key file:", err)
	}
	var key papernet.SigningKey
	err = json.Unmarshal(keyData, &key)
	if err != nil {
		log.Fatalln("could not read key file:", err)
	}

	googleOAuthClient, err := auth.NewGoogleClient(cfg.Auth.Google)
	if err != nil {
		log.Fatalln("could not read google oauth config:", err)
	}

	encoder := auth.EncodeDecoder{Key: key.Key}
	authenticator := auth.Authenticator{
		Decoder: &encoder,
		Store:   &userStore,
	}

	// Start web server
	addr := ":1705"
	server, err := gin.New(addr, authenticator)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	// Paper handler
	paperHandler := &web.PaperHandler{
		Store:     &paperStore,
		Index:     &index,
		TagIndex:  &tagIndex,
		UserStore: &userStore,
	}
	for _, route := range paperHandler.Routes() {
		server.Register(route)
	}

	// User handler
	userHandler := &web.UserHandler{
		Encoder:      &encoder,
		GoogleClient: googleOAuthClient,
		Store:        &userStore,
	}
	for _, route := range userHandler.Routes() {
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

	// Import handler
	teamHandler := &web.TeamHandler{
		Store:      &teamStore,
		PaperStore: &paperStore,
		UserStore:  &userStore,
	}
	for _, route := range teamHandler.Routes() {
		server.Register(route)
	}

	log.Println("server started, listening on", addr)
	log.Fatal(server.Start())
}
