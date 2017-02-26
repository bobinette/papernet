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
		"id":           paper.ID,
		"title":        paper.Title,
		"tags":         paper.Tags,
		"tags_keyword": paper.Tags,
		"authors":      paper.Authors,
		"arxivID":      paper.ArxivID,
	}

	return s.index.Index(strconv.Itoa(paper.ID), data)
}

func (s *PaperIndex) Delete(id int) error {
	return s.index.Delete(strconv.Itoa(id))
}

func (s *PaperIndex) Search(search papernet.PaperSearch) (papernet.PaperSearchResults, error) {
	q := andQ(
		query.NewMatchAllQuery(),
		s.searchQ(search.Q),
		s.searchIDs(search.IDs),
		s.termsQuery(search.ArxivIDs, "arxivID"),
		s.searchTags(search.Tags),
	)

	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.SortBy([]string{"id"})

	if search.Limit > 0 {
		searchRequest.Size = int(search.Limit)
	}
	searchRequest.From = int(search.Offset)

	tagsFacet := bleve.NewFacetRequest("tags_keyword", 10)
	searchRequest.AddFacet("tags", tagsFacet)

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

	facets := papernet.PaperSearchFacets{
		Tags: make(papernet.PaperSearchTagsFacet, len(searchResults.Facets["tags"].Terms)),
	}
	for i, term := range searchResults.Facets["tags"].Terms {
		facets.Tags[i].Tag = term.Term
		facets.Tags[i].Count = term.Count
	}

	return papernet.PaperSearchResults{
		IDs:    ids,
		Facets: facets,
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

func (s *PaperIndex) searchQ(queryString string) query.Query {
	words := strings.Fields(queryString)

	ands := make([]query.Query, 0, len(words))
	for _, word := range words {
		ands = append(ands, orQ(
			s.searchTitleQ(word),
			s.searchTagsQ(word),
			s.searchAuthorsQ(word),
		))
	}

	return andQ(ands...)
}

func (s *PaperIndex) searchTitleQ(queryString string) query.Query {
	analyzer := s.index.Mapping().AnalyzerNamed(en.AnalyzerName)
	tokens := analyzer.Analyze([]byte(queryString))
	if len(tokens) == 0 {
		return nil
	}

	conjuncts := make([]query.Query, len(tokens))
	for i, token := range tokens {
		conjuncts[i] = &query.PrefixQuery{
			Prefix: string(token.Term),
			Field:  "title",
		}
	}

	return query.NewConjunctionQuery(conjuncts)
}

func (s *PaperIndex) searchTagsQ(queryString string) query.Query {
	analyzer := s.index.Mapping().AnalyzerNamed(simple.Name)
	tokens := analyzer.Analyze([]byte(queryString))
	if len(tokens) == 0 {
		return nil
	}

	conjuncts := make([]query.Query, len(tokens))
	for i, token := range tokens {
		conjuncts[i] = &query.PrefixQuery{
			Prefix: string(token.Term),
			Field:  "tags",
		}
	}

	return query.NewConjunctionQuery(conjuncts)
}

func (s *PaperIndex) searchAuthorsQ(queryString string) query.Query {
	analyzer := s.index.Mapping().AnalyzerNamed(simple.Name)
	tokens := analyzer.Analyze([]byte(queryString))
	if len(tokens) == 0 {
		return nil
	}

	conjuncts := make([]query.Query, len(tokens))
	for i, token := range tokens {
		conjuncts[i] = &query.PrefixQuery{
			Prefix: string(token.Term),
			Field:  "authors",
		}
	}

	return query.NewConjunctionQuery(conjuncts)
}

func (*PaperIndex) searchIDs(ids []int) query.Query {
	docIDs := make([]string, len(ids))
	for i, id := range ids {
		docIDs[i] = strconv.Itoa(id)
	}
	return query.NewDocIDQuery(docIDs)
}

func (*PaperIndex) searchTags(tags []string) query.Query {
	if len(tags) == 0 {
		return nil
	}

	conjuncts := make([]query.Query, len(tags))
	for i, tag := range tags {
		conjuncts[i] = &query.MatchQuery{
			Match: tag,
			Field: "tags_keyword",
		}
	}

	return query.NewConjunctionQuery(conjuncts)
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
