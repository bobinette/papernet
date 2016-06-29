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
	dbpath = "data/papernet.bolt.db"
)

func main() {
	flag.StringVar(&dbpath, "dbpath", dbpath, "path to the db")
	flag.Parse()

	db, err := database.NewBoltDB(dbpath)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	defer db.Close()

	// graphDBPath := fmt.Sprintf("%s.cayley", dbpath)
	// err = graph.InitQuadStore("bolt", graphDBPath, graph.Options{"ignore_duplicate": true})
	// if err != nil && err != graph.ErrDatabaseExists {
	// 	log.Fatalf("could not init quadstore: %v", err)
	// }
	// g, err := cayley.NewGraph("bolt", graphDBPath, nil)
	// if err != nil {
	// 	log.Fatalf("could not open cayley: %v", err)
	// }
	// defer g.Close()

	r := gin.Default()
	h := papernet.NewHandler(db)
	h.Register(r)
	r.Run(":8080")
}
