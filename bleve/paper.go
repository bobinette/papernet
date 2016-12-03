package bleve

import (
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
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

func (s *PaperIndex) Delete(id int) error {
	return s.index.Delete(strconv.Itoa(id))
}

func (s *PaperIndex) Search(search papernet.PaperSearch) ([]int, error) {
	q := andQ(
		query.NewMatchAllQuery(),
		s.searchTitleOrTags(search.Q),
		s.searchIDs(search.IDs),
	)

	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.SortBy([]string{"id"})

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(searchResults.Hits))
	for i, hit := range searchResults.Hits {
		ids[i], err = strconv.Atoi(hit.ID)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func andQ(qs ...query.Query) query.Query {
	ands := make([]query.Query, 0, len(qs))
	for _, q := range qs {
		if q != nil {
			ands = append(ands, q)
		}
	}

	if len(ands) == 0 {
		return nil
	}
	return query.NewConjunctionQuery(ands)
}

func orQ(qs ...query.Query) query.Query {
	ors := make([]query.Query, 0, len(qs))
	for _, q := range qs {
		if q != nil {
			ors = append(ors, q)
		}
	}

	if len(ors) == 0 {
		return nil
	}
	return query.NewDisjunctionQuery(ors)
}

func (s *PaperIndex) searchTitleOrTags(queryString string) query.Query {
	words := strings.Fields(queryString)

	ands := make([]query.Query, 0, len(words))
	for _, word := range words {
		ands = append(ands, orQ(
			s.searchTitle(word),
			s.searchTags(word),
		))
	}

	return andQ(ands...)
}

func (s *PaperIndex) searchTitle(queryString string) query.Query {
	analyzer := s.index.Mapping().AnalyzerNamed(en.AnalyzerName)
	tokens := analyzer.Analyze([]byte(queryString))
	if len(tokens) == 0 {
		return nil
	}

	conjuncs := make([]query.Query, len(tokens))
	for i, token := range tokens {
		conjuncs[i] = &query.PrefixQuery{
			Prefix: string(token.Term),
			Field:  "title",
		}
	}

	return query.NewConjunctionQuery(conjuncs)
}

func (s *PaperIndex) searchTags(queryString string) query.Query {
	analyzer := s.index.Mapping().AnalyzerNamed(simple.Name)
	tokens := analyzer.Analyze([]byte(queryString))
	if len(tokens) == 0 {
		return nil
	}

	conjuncs := make([]query.Query, len(tokens))
	for i, token := range tokens {
		conjuncs[i] = &query.PrefixQuery{
			Prefix: string(token.Term),
			Field:  "tags",
		}
	}

	return query.NewConjunctionQuery(conjuncs)
}

func (*PaperIndex) searchIDs(ids []int) query.Query {
	if len(ids) == 0 {
		return nil
	}

	docIDs := make([]string, len(ids))
	for i, id := range ids {
		docIDs[i] = strconv.Itoa(id)
	}
	return query.NewDocIDQuery(docIDs)
}

// ------------------------------------------------------------------------------------------------
// Mapping
// ------------------------------------------------------------------------------------------------

func createMapping() mapping.IndexMapping {
	// a generic reusable mapping for english text -- from blevesearch/beer-search
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	simpleMapping := bleve.NewTextFieldMapping()
	simpleMapping.Analyzer = simple.Name

	// Paper mapping
	paperMapping := bleve.NewDocumentMapping()
	paperMapping.AddFieldMappingsAt("title", englishTextFieldMapping)
	paperMapping.AddFieldMappingsAt("tags", simpleMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = paperMapping
	return indexMapping
}
