package imports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bobinette/papernet/clients/paper"
	"github.com/bobinette/papernet/errors"
)

type mockMapping struct {
	mapping map[int]map[string]map[string]int
}

func (m *mockMapping) Save(userID, paperID int, source, ref string) error { return nil }
func (m *mockMapping) Get(userID int, source, ref string) (int, error) {
	return m.mapping[userID][source][ref], nil
}

func insertPaper(w http.ResponseWriter, req *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]map[string]int{
		"data": map[string]int{"id": 12},
	})
}

func mockPaperService(t *testing.T) *paper.Client {
	srv := httptest.NewServer(http.HandlerFunc(insertPaper))
	return paper.NewClient(&http.Client{}, srv.URL)
}

type mockImporter struct {
	source  string
	results SearchResults

	calls []struct {
		q      string
		limit  int
		offset int
	}
}

func (m *mockImporter) Source() string { return m.source }
func (m *mockImporter) Search(ctx context.Context, q string, limit, offset int) (SearchResults, error) {
	return m.results, nil
}

func TestSearchService_Search(t *testing.T) {
	searcher1 := &mockImporter{
		source: "searcher 1",
		results: SearchResults{
			Papers: []Paper{
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
			Pagination: Pagination{
				Limit:  2,
				Offset: 0,
				Total:  4,
			},
		},
	}

	searcher2 := &mockImporter{
		source: "searcher 2",
		results: SearchResults{
			Papers: []Paper{
				{
					Reference: "Reference 1",
					Title:     "Title 1",
					Summary:   "Summary 1",
					Tags:      []string{"Tags 1"},
					Authors:   []string{"Authors 1"},
				},
			},
			Pagination: Pagination{
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

	client := mockPaperService(t)
	service := NewService(mapping, client, searcher1, searcher2)
	for name, tt := range tts {
		res, err := service.Search(context.Background(), userID, "", 2, 0, tt.sources)
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
		searchers []Searcher
		sources   []string
	}{
		"no searchers": {
			searchers: nil,
			sources:   []string{"source"},
		},
		"one searcher": {
			searchers: []Searcher{
				&mockImporter{source: "searcher"},
			},
			sources: []string{"source"},
		},
		"several searchers": {
			searchers: []Searcher{
				&mockImporter{source: "searcher 1"},
				&mockImporter{source: "searcher 2"},
			},
			sources: []string{"source"},
		},
	}

	for name, tt := range tts {
		client := mockPaperService(t)
		service := NewService(&mockMapping{}, client, tt.searchers...)
		_, err := service.Search(context.Background(), 1, "q", 20, 0, tt.sources)
		if assert.Error(t, err, name) {
			errors.AssertCode(t, err, http.StatusBadRequest)
		}
	}
}
