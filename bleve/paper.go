package bleve

import (
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/search/query"

	"github.com/bobinette/papernet"
)

type PaperIndex struct {
	index bleve.Index
}

func (s *PaperIndex) Open(path string) error {
	index, err := bleve.Open(path)
	if err != nil {
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
		"title":   paper.Title,
		"tags":    paper.Tags,
		"arxivID": paper.ArxivID,
	}

	return s.index.Index(strconv.Itoa(paper.ID), data)
}

func (s *PaperIndex) Delete(id int) error {
	return s.index.Delete(strconv.Itoa(id))
}

func (s *PaperIndex) Search(search papernet.PaperSearch) (papernet.PaperSearchResults, error) {
	q := andQ(
		query.NewMatchAllQuery(),
		s.searchTitleOrTags(search.Q),
		s.searchIDs(search.IDs),
		s.termsQuery(search.ArxivIDs, "arxivID"),
	)

	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.SortBy([]string{"_id"})

	if search.Limit > 0 {
		searchRequest.Size = int(search.Limit)
	}
	searchRequest.From = int(search.Offset)

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return papernet.PaperSearchResults{}, err
	}

	ids := make([]int, len(searchResults.Hits))
	for i, hit := range searchResults.Hits {
		ids[i], err = strconv.Atoi(hit.ID)
		if err != nil {
			return papernet.PaperSearchResults{}, err
		}
	}

	return papernet.PaperSearchResults{
		IDs: ids,
		Pagination: papernet.Pagination{
			Total:  searchResults.Total,
			Limit:  search.Limit,
			Offset: search.Offset,
		},
	}, nil
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
	docIDs := make([]string, len(ids))
	for i, id := range ids {
		docIDs[i] = strconv.Itoa(id)
	}
	return query.NewDocIDQuery(docIDs)
}

func (*PaperIndex) termsQuery(terms []string, field string) query.Query {
	if len(terms) == 0 {
		return nil
	}

	ors := make([]query.Query, len(terms))
	for i, term := range terms {
		ors[i] = &query.TermQuery{
			Term:  term,
			Field: field,
		}
	}

	return orQ(ors...)
}
