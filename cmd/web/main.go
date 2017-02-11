package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
	"github.com/bobinette/papernet/web"
)

func main() {
	// Load signing key
	data, err := ioutil.ReadFile("hs256.json")
	if err != nil {
		log.Fatalln("could not open key file:", err)
	}
	var key papernet.SigningKey
	err = json.Unmarshal(data, &key)
	if err != nil {
		log.Fatalln("could not read key file:", err)
	}

	// Create repositories
	driver := bolt.Driver{}
	defer driver.Close()
	err = driver.Open("data/papernet.db")
	if err != nil {
		log.Fatalln("could not open db:", err)
	}

	paperStore := bolt.PaperStore{Driver: &driver}
	tagIndex := bolt.TagIndex{Driver: &driver}
	userRepo := bolt.UserRepository{Driver: &driver}

	// Create index
	index := bleve.PaperIndex{}
	err = index.Open("data/papernet.index")
	defer index.Close()
	if err != nil {
		log.Fatalln("could not open index:", err)
	}

	// Auth
	googleOAuthClient, err := auth.NewGoogleClient("oauth_google.json")
	if err != nil {
		log.Fatalln("could not read google oauth config:", err)
	}

	encoder := auth.EncodeDecoder{Key: key.Key}

	// Start web server
	handler, err := gin.New(&paperStore, &index, &tagIndex, &userRepo, key, googleOAuthClient)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	// Paper handler
	paperHandler := &web.PaperHandler{
		Store:     &paperStore,
		Index:     &index,
		TagIndex:  &tagIndex,
		UserStore: &userRepo,
	}
	for _, route := range paperHandler.Routes() {
		handler.Register(route)
	}

	// User handler
	userHandler := &web.UserHandler{
		Encoder:      &encoder,
		GoogleClient: googleOAuthClient,
		Store:        &userRepo,
	}
	for _, route := range userHandler.Routes() {
		handler.Register(route)
	}

	addr := ":1705"
	log.Println("server started, listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
