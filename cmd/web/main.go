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
	// Create repository
	repo := bolt.PaperRepository{}
	err := repo.Open("data/papernet.db")
	defer repo.Close()
	if err != nil {
		log.Fatalln("could not open db:", err)
	}

	// Create index
	index := bleve.PaperSearch{}
	err = index.Open("data/papernet.index")
	defer index.Close()
	if err != nil {
		log.Fatalln("could not open index:", err)
	}

	// Start web server
	handler, err := gin.New(&repo, &index)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	addr := ":1705"
	log.Println("server started, listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
