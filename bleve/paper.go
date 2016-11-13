package bleve

import (
	"encoding/binary"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"

	"github.com/bobinette/papernet"
)

type PaperSearch struct {
	Repository papernet.PaperRepository

	index bleve.Index
}

func (s *PaperSearch) Open(path string) error {
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		indexMapping := createMapping()
		index, err = bleve.New(path, indexMapping)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	s.index = index
	return nil
}

func (s *PaperSearch) Close() error {
	if s.index == nil {
		return nil
	}

	return s.index.Close()
}

func (s *PaperSearch) Index(paper *papernet.Paper) error {
	data := map[string]interface{}{
		"title": paper.Title,
	}

	return s.index.Index(strconv.Itoa(paper.ID), data)
}

func (s *PaperSearch) Search(titlePrefix string) ([]int, error) {
	var q query.Query
	if titlePrefix != "" {
		titleQuery := query.NewPrefixQuery(titlePrefix)
		titleQuery.Field = "title"
		q = titleQuery
	} else {
		q = query.NewMatchAllQuery()
	}

	search := bleve.NewSearchRequest(q)
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

// ------------------------------------------------------------------------------------------------
// Mapping
// ------------------------------------------------------------------------------------------------

func createMapping() mapping.IndexMapping {
	// a generic reusable mapping for english text -- from blevesearch/beer-search
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	// Paper mapping
	paperMapping := bleve.NewDocumentMapping()
	paperMapping.AddFieldMappingsAt("title", englishTextFieldMapping)
	paperMapping.AddFieldMappingsAt("version", bleve.NewNumericFieldMapping())

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("paper", paperMapping)
	return indexMapping
}

// ------------------------------------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------------------------------------

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
