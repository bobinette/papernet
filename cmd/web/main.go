package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
	"github.com/bobinette/papernet/oauth"
)

func main() {
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

	paperRepo := bolt.PaperRepository{Driver: &driver}
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
	googleOAuthClient, err := oauth.NewGoogleOAuthClient("oauth_google.json")
	if err != nil {
		log.Fatalln("could not read google oauth config:", err)
	}

	// Start web server
	handler, err := gin.New(&paperRepo, &index, &tagIndex, &userRepo, key, googleOAuthClient)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	addr := ":1705"
	log.Println("server started, listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
