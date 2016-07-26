package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/bobinette/papernet/web"
)

var (
	dbPath    = "data/papernet.bolt.db"
	indexPath = "data/papernet.bleve"
)

func main() {
	flag.StringVar(&dbPath, "dbPath", dbPath, "path to the db")
	flag.StringVar(&indexPath, "indexPath", indexPath, "path to the search store")
	flag.Parse()

	srv, err := web.NewServer(dbPath, indexPath)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	http.ListenAndServe(":8080", srv)
}
