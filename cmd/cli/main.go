package main

import (
	"flag"
	"log"
	"os"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/etl"
)

type repoResult struct {
	repo papernet.PaperRepository
	f    func()
}

func createRepository(addr string) (*repoResult, error) {
	driver := bolt.Driver{}

	err := driver.Open(addr)
	if err != nil {
		return nil, err
	}

	repo := bolt.PaperRepository{Driver: &driver}
	return &repoResult{
		repo: &repo,
		f:    func() { driver.Close() },
	}, nil
}

type indexResult struct {
	index papernet.PaperIndex
	f     func()
}

func createIndex(addr string) (*indexResult, error) {
	index := bleve.PaperIndex{}
	err := index.Open("data/papernet.index")
	if err != nil {
		return nil, err
	}

	return &indexResult{
		index: &index,
		f:     func() { index.Close() },
	}, nil
}

func parse(resource string) error {
	log.Println("Importing", resource)
	repoRes, err := createRepository("data/papernet.db")
	if err != nil {
		return err
	}
	defer repoRes.f()

	indexRes, err := createIndex("data/papernet.index")
	if err != nil {
		return err
	}
	defer indexRes.f()

	importer := etl.Importer{}

	paper, err := importer.Import(resource)
	if err != nil {
		return err
	}

	paper.ID = 19
	// Save the paper
	err = repoRes.repo.Upsert(&paper)
	if err != nil {
		return err
	}

	err = indexRes.index.Index(&paper)
	if err != nil {
		return err
	}

	log.Println("Done. Paper ID:", paper.ID)
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

	// Do stuff
	switch os.Args[1] {
	case "reindex":
		// Create repository
		driver := bolt.Driver{}
		defer driver.Close()
		err := driver.Open("data/papernet.db")
		if err != nil {
			log.Fatalln("could not open db:", err)
		}
		repo := bolt.PaperRepository{Driver: &driver}

		// Create index
		index := bleve.PaperIndex{}
		err = index.Open("data/papernet.index")
		defer index.Close()
		if err != nil {
			log.Fatalln("could not open index:", err)
		}

		log.Println("Reindexing all the papers")
		err = reindexAll(&repo, &index, *verbose)
		if err != nil {
			log.Fatal("error reindexing:", err)
		}
		log.Println("Done")
	case "parse":
		if len(os.Args) < 3 {
			log.Fatalln("missing url to parse")
		}
		err := parse(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalf(`unknown argument %s. Should be one of:
  - reindex
  - parse
 `, os.Args[1])
	}
}

func reindexAll(repo papernet.PaperRepository, index papernet.PaperIndex, v bool) error {
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
