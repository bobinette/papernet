package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/imports"
)

type mockMapping struct {
	mapping map[int]map[string]map[string]int
}

func (m *mockMapping) Save(userID, paperID int, source, ref string) error { return nil }
func (m *mockMapping) Get(userID int, source, ref string) (int, error) {
	return m.mapping[userID][source][ref], nil
}

type mockImporter struct {
	source  string
	results imports.SearchResults

	calls []struct {
		q      string
		limit  int
		offset int
	}
}

func (m *mockImporter) Source() string { return m.source }
func (m *mockImporter) Search(q string, limit, offset int, ctx context.Context) (imports.SearchResults, error) {
	return m.results, nil
}
func (m *mockImporter) Import(ref string, ctx context.Context) (imports.Paper, error) {
	return imports.Paper{}, nil
}

func TestSearchService_Search(t *testing.T) {
	searcher1 := &mockImporter{
		source: "searcher 1",
		results: imports.SearchResults{
			Papers: []imports.Paper{
				{
					Reference: "Reference 1",
					Title:     "Title 1",
					Summary:   "Summary 1",
					Tags:      []string{"Tags 1"},
					Authors:   []string{"Authors 1"},
				},
				{
					Reference: "Reference 2",
					Title:     "Title 2",
					Summary:   "Summary 2",
					Tags:      []string{"Tags 2"},
					Authors:   []string{"Authors 2"},
				},
			},
			Pagination: imports.Pagination{
				Limit:  2,
				Offset: 0,
				Total:  4,
			},
		},
	}

	searcher2 := &mockImporter{
		source: "searcher 2",
		results: imports.SearchResults{
			Papers: []imports.Paper{
				{
					Reference: "Reference 1",
					Title:     "Title 1",
					Summary:   "Summary 1",
					Tags:      []string{"Tags 1"},
					Authors:   []string{"Authors 1"},
				},
			},
			Pagination: imports.Pagination{
				Limit:  2,
				Offset: 0,
				Total:  1,
			},
		},
	}

	userID := 1
	mapping := &mockMapping{
		mapping: map[int]map[string]map[string]int{
			userID: map[string]map[string]int{
				"searcher 1": map[string]int{
					"Reference 1": 1,
				},
				"searcher 2": map[string]int{
					"Reference 1": 2,
				},
			},
		},
	}

	tts := map[string]struct {
		sources []string
		res     map[string][]struct {
			id  int
			ref string
		}
	}{
		"no source specified": {
			sources: nil,
			res: map[string][]struct {
				id  int
				ref string
			}{
				"searcher 1": {
					{
						id:  1,
						ref: "Reference 1",
					},
					{
						id:  0,
						ref: "Reference 2",
					},
				},
				"searcher 2": {
					{
						id:  2,
						ref: "Reference 1",
					},
				},
			},
		},
		"only searcher 1": {
			sources: []string{"searcher 1"},
			res: map[string][]struct {
				id  int
				ref string
			}{
				"searcher 1": {
					{
						id:  1,
						ref: "Reference 1",
					},
					{
						id:  0,
						ref: "Reference 2",
					},
				},
			},
		},
	}

	service := NewImportService(mapping, searcher1, searcher2)
	for name, tt := range tts {
		res, err := service.Search(userID, "", 2, 0, tt.sources, context.Background())
		assert.NoError(t, err, name)

		for source, expected := range tt.res {
			actual := res[source]
			if assert.Equal(t, len(expected), len(actual.Papers), "%s - source: %s - len", name, source) {

				for i, e := range expected {
					a := actual.Papers[i]
					assert.Equal(t, e.id, a.ID, "%s - source: %s - id", name, source)
					assert.Equal(t, e.ref, a.Reference, "%s - source: %s - ref", name, source)
				}
			}
		}
	}
}

func TestSearchService_Search_UnknownSource(t *testing.T) {
	tts := map[string]struct {
		searchers []imports.Importer
		sources   []string
	}{
		"no searchers": {
			searchers: nil,
			sources:   []string{"source"},
		},
		"one searcher": {
			searchers: []imports.Importer{
				&mockImporter{source: "searcher"},
			},
			sources: []string{"source"},
		},
		"several searchers": {
			searchers: []imports.Importer{
				&mockImporter{source: "searcher 1"},
				&mockImporter{source: "searcher 2"},
			},
			sources: []string{"source"},
		},
	}

	for name, tt := range tts {
		service := NewImportService(&mockMapping{}, tt.searchers...)
		_, err := service.Search(1, "q", 20, 0, tt.sources, context.Background())
		if assert.Error(t, err, name) {
			errors.AssertCode(t, err, http.StatusBadRequest)
		}
	}
}
