package main

import (
	"flag"
	"log"
	"os"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	// Create repository
	driver := bolt.Driver{}
	defer driver.Close()
	err := driver.Open("data/papernet.db")
	if err != nil {
		log.Fatalln("could not open db:", err)
	}
	repo := bolt.PaperRepository{Driver: &driver}

	// Create index
	index := bleve.PaperSearch{}
	err = index.Open("data/papernet.index")
	defer index.Close()
	if err != nil {
		log.Fatalln("could not open index:", err)
	}

	if len(os.Args) < 2 {
		log.Fatal(`argument missing. One of:
  - reindex
 `)
	}

	// Do stuff
	switch os.Args[len(os.Args)-1] {
	case "reindex":
		log.Println("Reindexing all the papers")
		err := reindexAll(&repo, &index, *verbose)
		if err != nil {
			log.Fatal("error reindexing:", err)
		}
		log.Println("Done")
	default:
		log.Fatalf(`unknown argument %s. Should be one of:
  - reindex
 `, os.Args[1])
	}
}

func reindexAll(repo papernet.PaperRepository, index papernet.PaperSearch, v bool) error {
	papers, err := repo.List()
	if err != nil {
		return err
	}

	for _, paper := range papers {
		err := index.Index(paper)
		if err != nil {
			return err
		}

		if v {
			log.Println("Paper", paper.ID, "reindexed")
		}
	}

	return nil
}
