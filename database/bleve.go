package database

import (
	"log"
	"strconv"

	"github.com/blevesearch/bleve"

	"github.com/bobinette/papernet/models"
)

type bleveSearch struct {
	index bleve.Index
}

func NewBleveSearch(searchpath string) (Search, error) {
	index, err := bleve.Open(searchpath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Print("Creating index...")
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(searchpath, mapping)
		log.Println("Done")
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &bleveSearch{
		index: index,
	}, nil
}

func (s *bleveSearch) Find(q string) ([]int, error) {
	query := bleve.NewMatchQuery(q)
	search := bleve.NewSearchRequest(query)
	searchResults, err := s.index.Search(search)
	if err != nil {
		return nil, err
	}

	ids := make([]int, searchResults.Total)
	for i, hit := range searchResults.Hits {
		ids[i], err = strconv.Atoi(hit.ID)
		if err != nil {
			return nil, err
		}
		log.Printf("%+v", hit)
	}
	return ids, nil
}

func (s *bleveSearch) Index(p *models.Paper) error {
	data := map[string]interface{}{
		"title": string(p.Title),
		"tags":  p.Tags,
	}

	// Index title
	err := s.index.Index(strconv.Itoa(p.ID), data)
	if err != nil {
		return err
	}

	// Index tags
	// err = s.index.Index(strconv.Itoa(p.ID), p.Tags)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *bleveSearch) Close() error {
	return s.index.Close()
}
