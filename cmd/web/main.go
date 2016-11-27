package main

import (
	"log"
	"net/http"

	// "github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
)

func main() {
	// Create repositories
	driver := bolt.Driver{}
	defer driver.Close()
	err := driver.Open("data/papernet.db")
	if err != nil {
		log.Fatalln("could not open db:", err)
	}

	paperRepo := bolt.PaperRepository{Driver: &driver}
	tagIndex := bolt.TagIndex{Driver: &driver}

	// Create index
	index := bleve.PaperIndex{}
	err = index.Open("data/papernet.index")
	defer index.Close()
	if err != nil {
		log.Fatalln("could not open index:", err)
	}

	// Start web server
	handler, err := gin.New(&paperRepo, &index, &tagIndex)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	addr := ":1705"
	log.Println("server started, listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
