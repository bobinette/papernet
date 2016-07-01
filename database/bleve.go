package database

import (
	"log"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/language/en"

	"github.com/bobinette/papernet/models"
)

type bleveSearch struct {
	index bleve.Index
}

func NewBleveSearch(searchpath string) (Search, error) {
	index, err := bleve.Open(searchpath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		em := bleve.NewTextFieldMapping()
		em.Analyzer = en.AnalyzerName
		dm := bleve.NewDocumentMapping()
		dm.AddFieldMappingsAt("title", em)

		mapping := bleve.NewIndexMapping()
		mapping.AddDocumentMapping("paper", dm)
		index, err = bleve.New(searchpath, mapping)
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
	return s.index.Index(strconv.Itoa(p.ID), string(p.Title))
}

func (s *bleveSearch) Close() error {
	return s.index.Close()
}
