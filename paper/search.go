package paper

import (
	"log"
	"strconv"

	"github.com/blevesearch/bleve"
)

type Search struct {
	index bleve.Index
}

func NewSearch(indexPath string) (*Search, error) {
	index, err := bleve.Open(indexPath)

	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Print("Creating index...")
		mapping := createMapping()
		index, err = bleve.New(indexPath, mapping)
		log.Println("Done")
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &Search{
		index: index,
	}, nil
}

func (s *Search) Find(q string) ([]int, error) {
	query := bleve.NewPrefixQuery(q)
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
	}
	return ids, nil
}

func (s *Search) Index(p *Paper) error {
	data := map[string]interface{}{
		"title": string(p.Title),
		"tags":  p.Tags,
	}

	err := s.index.Index(strconv.Itoa(p.ID), data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Search) Close() error {
	return s.index.Close()
}

// ------------------------------------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------------------------------------

func createMapping() *bleve.IndexMapping {
	return bleve.NewIndexMapping()
}
