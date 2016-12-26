package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/etl"
	"github.com/bobinette/papernet/etl/crawlers"
	"github.com/bobinette/papernet/etl/scrapers"
)

type storeResult struct {
	store papernet.PaperStore
	f     func()
}

func createStore(addr string) (papernet.PaperStore, func(), error) {
	driver := bolt.Driver{}

	err := driver.Open(addr)
	if err != nil {
		return nil, func() {}, err
	}

	store := bolt.PaperStore{Driver: &driver}
	return &store, func() { driver.Close() }, nil
}

type indexResult struct {
	index papernet.PaperIndex
	f     func()
}

func createIndex(addr string) (papernet.PaperIndex, func(), error) {
	index := bleve.PaperIndex{}
	err := index.Open(addr)
	if err != nil {
		return nil, func() {}, err
	}

	return &index, func() { index.Close() }, nil
}

func parse(resource string, store papernet.PaperStore, index papernet.PaperIndex) error {
	log.Println("Importing", resource)
	importer := etl.Importer{}
	crawler, ok := crawlers.New("html")
	if !ok {
		return fmt.Errorf("no crawler for %s", "html")
	}

	scraper, ok := scrapers.New("arxiv")
	if !ok {
		return fmt.Errorf("no scraper for %s", "html")
	}

	papers, err := importer.Import(resource, crawler, scraper)
	if err != nil {
		return err
	}

	for _, paper := range papers {
		// Save the paper
		// err = store.Upsert(&paper)
		// if err != nil {
		// 	return err
		// }

		// err = index.Index(&paper)
		// if err != nil {
		// 	return err
		// }

		log.Println(paper)
		log.Println("Done. Paper ID:", paper.ID)
	}

	log.Printf("Done. %d papers added.", len(papers))
	return nil
}

func reindexAll(store papernet.PaperStore, index papernet.PaperIndex, v bool) error {
	papers, err := store.List()
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

func restoreDates(store papernet.PaperStore) error {
	papers, err := store.List()
	if err != nil {
		return err
	}

	nilTime := time.Time{}
	for _, paper := range papers {
		if paper.CreatedAt.Equal(nilTime) {
			paper.CreatedAt = time.Now()
		}
		if paper.UpdatedAt.Equal(nilTime) {
			paper.UpdatedAt = time.Now()
		}

		err = store.Upsert(paper)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatal(`argument missing. One of:
  - reindex
  - parse <url>
 `)
	}

	store, f, err := createStore("data/papernet.db")
	defer f()
	if err != nil {
		log.Fatalln(err)
	}

	index, f, err := createIndex("data/papernet.index")
	defer f()
	if err != nil {
		log.Fatalln(err)
	}

	// Do stuff
	switch os.Args[1] {
	case "reindex":
		log.Println("Reindexing all the papers")
		err = reindexAll(store, index, *verbose)
		if err != nil {
			log.Fatal("error reindexing:", err)
		}
		log.Println("Done")
	case "parse":
		if len(os.Args) < 3 {
			log.Fatalln("missing url to parse")
		}
		err := parse(os.Args[2], store, index)
		if err != nil {
			log.Fatalln(err)
		}
	case "restore-dates":
		if err := restoreDates(store); err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalf(`unknown argument %s. Should be one of:
  - reindex
  - parse
 `, os.Args[1])
	}
}
