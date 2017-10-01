package arxiv

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bobinette/papernet/imports"
)

func TestImporter_Search(t *testing.T) {
	importer := NewSearcher()

	data, err := ioutil.ReadFile("yolo_search.xml")
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(data))
	}))
	defer ts.Close()

	apiURLStr = ts.URL
	res, err := importer.Search(context.Background(), "YOLO", 2, 1)
	assert.NoError(t, err)

	assert.Equal(t, imports.Pagination{Limit: 2, Offset: 1, Total: 7}, res.Pagination)

	if assert.Equal(t, 2, len(res.Papers)) {
		refs := []string{"1705.09587", "1705.05922"}

		for i, ref := range refs {
			paper := res.Papers[i]
			assert.Equal(t, ref, paper.Reference)
			assert.Equal(t, "arxiv", paper.Source)
		}
	}
}

func TestCraftURL(t *testing.T) {
	tts := map[string]struct {
		q        string
		limit    int
		offset   int
		expected map[string][]string
	}{
		"no q": {
			q:      "",
			limit:  10,
			offset: 20,
			expected: map[string][]string{
				"max_results": []string{"10"},
				"start":       []string{"20"},
				"sortBy":      []string{"submittedDate"},
				"sortOrder":   []string{"descending"},
			},
		},
		"one word q": {
			q:      "YOLO",
			limit:  5,
			offset: 0,
			expected: map[string][]string{
				"search_query": []string{"all:YOLO"},
				"max_results":  []string{"5"},
				"start":        []string{"0"},
				"sortBy":       []string{"submittedDate"},
				"sortOrder":    []string{"descending"},
			},
		},
		"multi word q": {
			q:      "YOLO X9 44 abcd",
			limit:  89,
			offset: 143,
			expected: map[string][]string{
				"search_query": []string{"all:YOLO AND X9 AND 44 AND abcd"},
				"max_results":  []string{"89"},
				"start":        []string{"143"},
				"sortBy":       []string{"submittedDate"},
				"sortOrder":    []string{"descending"},
			},
		},
	}

	for name, tt := range tts {
		u := craftURL(tt.q, tt.limit, tt.offset)
		qp := u.Query()
		assert.Equal(t, len(tt.expected), len(qp), "%s - len", name)
		for k, expected := range tt.expected {
			actual := qp[k]
			assert.Equal(t, expected, actual, "%s - %s", name, k)
		}
	}
}

func TestCraftRefURL(t *testing.T) {
	tts := map[string]struct {
		ref      string
		expected map[string][]string
	}{
		"ref": {
			ref: "ref",
			expected: map[string][]string{
				"id_list": []string{"ref"},
			},
		},
	}

	for name, tt := range tts {
		u := craftRefURL(tt.ref)
		qp := u.Query()
		assert.Equal(t, len(tt.expected), len(qp), name)
		for k, expected := range tt.expected {
			actual := qp[k]
			assert.Equal(t, expected, actual, "%s - %s", name, k)
		}
	}
}

func TestExtractReference(t *testing.T) {
	tts := map[string]struct {
		input string
		ref   string
	}{
		"abstract": {
			input: "http://arxiv.org/abs/1234.5678v5",
			ref:   "1234.5678",
		},
		"pdf": {
			input: "http://arxiv.org/pdf/1234.5678v2",
			ref:   "1234.5678",
		},
		"physics...": {
			input: "http://arxiv.org/abs/quant-ph/1234.5678v1",
			ref:   "1234.5678",
		},
		"still extracts something even if invalid id": {
			input: "http://arxiv.org/abs/not-an-id",
			ref:   "not-an-id",
		},
	}

	for name, tt := range tts {
		ref := extractReference(tt.input)
		assert.Equal(t, tt.ref, ref, name)
	}
}
