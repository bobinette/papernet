package bleve

import (
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"

	"github.com/bobinette/papernet"
)

type PaperIndex struct {
	Repository papernet.PaperRepository

	index bleve.Index
}

func (s *PaperIndex) Open(path string) error {
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

func (s *PaperIndex) Close() error {
	if s.index == nil {
		return nil
	}

	return s.index.Close()
}

func (s *PaperIndex) Index(paper *papernet.Paper) error {
	data := map[string]interface{}{
		"title": paper.Title,
		"tags":  paper.Tags,
	}

	return s.index.Index(strconv.Itoa(paper.ID), data)
}

func (s *PaperIndex) Search(titlePrefix string) ([]int, error) {
	var q query.Query
	if titlePrefix != "" {
		tokens := strings.Fields(titlePrefix)
		titleConjuncts := make([]query.Query, len(tokens))
		tagConjuncs := make([]query.Query, len(tokens))
		for i, token := range tokens {
			titleConjuncts[i] = &query.PrefixQuery{
				Prefix: token,
				Field:  "title",
			}

			tagConjuncs[i] = &query.PrefixQuery{
				Prefix: token,
				Field:  "tags",
			}
		}
		q = query.NewDisjunctionQuery([]query.Query{
			query.NewConjunctionQuery(titleConjuncts),
			query.NewConjunctionQuery(tagConjuncs),
		})
	} else {
		q = query.NewMatchAllQuery()
	}

	search := bleve.NewSearchRequest(q)
	search.SortBy([]string{"id"})
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
	paperMapping.AddFieldMappingsAt("tags", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = paperMapping
	return indexMapping
}
