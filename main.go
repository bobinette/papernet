package main

import (
	"flag"
	// "fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet/database"
	"github.com/bobinette/papernet/web"
)

var (
	dbpath     = "data/papernet.bolt.db"
	searchpath = "data/papernet.bleve"
)

func main() {
	flag.StringVar(&dbpath, "dbpath", dbpath, "path to the db")
	flag.StringVar(&searchpath, "searchpath", searchpath, "path to the search store")
	flag.Parse()

	db, err := database.NewBoltDB(dbpath)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	defer db.Close()

	search, err := database.NewBleveSearch(searchpath)
	if err != nil {
		log.Fatalf("error connecting to search index: %v", err)
	}
	defer search.Close()

	r := gin.Default()
	h := papernet.NewHandler(db, search)
	h.Register(r)
	r.Run(":8081")
}
