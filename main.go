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

	// b, err := bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	// if err != nil {
	// 	log.Fatalf("open db: %v", err)
	// }
	// defer b.Close()

	// // Check buckets
	// err = b.Update(func(tx *bolt.Tx) error {
	// 	_, err := tx.CreateBucketIfNotExists([]byte("papers"))
	// 	if err != nil {
	// 		return fmt.Errorf("create bucket: %s", err)
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	log.Fatalf("db check: %v", err)
	// }

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

	db, err := database.NewBoltDB(dbpath)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	defer db.Close()

	r := gin.Default()
	h := papernet.NewHandler(db)
	h.Register(r)
	r.Run(":8080")
}
