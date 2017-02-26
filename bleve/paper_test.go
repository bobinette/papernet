package bleve

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"

	"github.com/bobinette/papernet"
)

func createIndex(t *testing.T) (*PaperIndex, func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	data, err := ioutil.ReadFile("mapping.json")
	if err != nil {
		t.Fatal(err)
	}

	var m mapping.IndexMappingImpl
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatal(err)
	}

	index, err := bleve.New(dir, &m)
	if err != nil {
		t.Fatal("error creating index", err)
	}

	if err != nil {
		os.RemoveAll(dir)
		t.Fatal("could not create bucket: ", err)
	}

	return &PaperIndex{index: index}, func() {
		if err := index.Close(); err != nil {
			t.Log(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Log(err)
		}
	}
}

func TestFind(t *testing.T) {
	index, f := createIndex(t)
	defer f()

	papers := []*papernet.Paper{
		&papernet.Paper{ID: 1, Title: "Title 1", Tags: []string{"tag"}},
		&papernet.Paper{ID: 2, Title: "Pizza yolo", Tags: []string{"tag"}},
		&papernet.Paper{ID: 3, Title: "titi et rominet", Tags: []string{"tag", "tech"}},
		&papernet.Paper{ID: 4, Title: "pizza", Tags: []string{"tag", "technique"}},
		&papernet.Paper{ID: 5, Title: "reinforcement learning", Tags: []string{"machine learning"}},
		&papernet.Paper{ID: 6, Title: "monte carlo", Tags: []string{"machine learning"}},
		&papernet.Paper{ID: 7, Title: "pizza yolo", Tags: []string{"tag", "robbery"}},
		&papernet.Paper{ID: 8, Title: "learning to build a machine", Tags: []string{"skillz", "DIY"}},
		&papernet.Paper{ID: 11, Title: "later that day", Tags: []string{"tag"}},
		&papernet.Paper{ID: 24, Title: "twenty four", Tags: []string{"tag", "24"}},
	}
	ids := make([]int, len(papers))
	for i, paper := range papers {
		if err := index.Index(paper); err != nil {
			t.Fatal("error indexing", paper.ID, err)
		}
		ids[i] = paper.ID
	}

	var tts = map[string]struct {
		Search   papernet.PaperSearch
		Expected papernet.PaperSearchResults
	}{
		"match all": {
			Search: papernet.PaperSearch{
				Q:     "",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: ids,
				Pagination: papernet.Pagination{
					Total:  uint64(len(ids)),
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"one word": {
			Search: papernet.PaperSearch{
				Q:     "pizza",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{2, 4, 7},
				Pagination: papernet.Pagination{
					Total:  3,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"partial word": {
			Search: papernet.PaperSearch{
				Q:     "ti",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{1, 3},
				Pagination: papernet.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"two words": {
			Search: papernet.PaperSearch{
				Q:     "pizza yolo",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{2, 7},
				Pagination: papernet.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"long words": {
			Search: papernet.PaperSearch{
				Q:     "reinforcement learning",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{5},
				Pagination: papernet.Pagination{
					Total:  1,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"long words spelling": {
			Search: papernet.PaperSearch{
				Q:     "mysuperlnog",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{},
				Pagination: papernet.Pagination{
					Total:  0,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"trailing space": {
			Search: papernet.PaperSearch{
				Q:     "titi ",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{3},
				Pagination: papernet.Pagination{
					Total:  1,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"by tags": {
			Search: papernet.PaperSearch{
				Q:     "tech",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{3, 4},
				Pagination: papernet.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"ro": {
			Search: papernet.PaperSearch{
				Q:     "pi yo ro",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{7},
				Pagination: papernet.Pagination{
					Total:  1,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"with uppercase letters": {
			Search: papernet.PaperSearch{
				Q:     "Learning",
				IDs:   ids,
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{5, 6, 8},
				Pagination: papernet.Pagination{
					Total:  3,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"by ids": {
			Search: papernet.PaperSearch{
				IDs:   []int{1, 3, 17},
				Limit: 10,
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{1, 3},
				Pagination: papernet.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"with q tag": {
			Search: papernet.PaperSearch{
				IDs:   ids,
				Limit: 10,
				Q:     "machine learning",
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{5, 6, 8},
				Pagination: papernet.Pagination{
					Total:  3,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"with fixed tag": {
			Search: papernet.PaperSearch{
				IDs:   ids,
				Limit: 10,
				Tags:  []string{"machine learning"},
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{5, 6},
				Pagination: papernet.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
		},
		"match all with limit": {
			Search: papernet.PaperSearch{
				Q:     "",
				IDs:   ids,
				Limit: 5,
			},
			Expected: papernet.PaperSearchResults{
				IDs: ids[:5],
				Pagination: papernet.Pagination{
					Total:  uint64(len(ids)),
					Limit:  5,
					Offset: 0,
				},
			},
		},
		"tag + order by id desc": {
			Search: papernet.PaperSearch{
				Tags:    []string{"tag"},
				IDs:     ids,
				Limit:   uint64(len(ids)),
				OrderBy: "-id",
			},
			Expected: papernet.PaperSearchResults{
				IDs: []int{24, 11, 7, 4, 3, 2, 1},
				Pagination: papernet.Pagination{
					Total:  7,
					Limit:  uint64(len(ids)),
					Offset: 0,
				},
			},
		},
	}

	for name, tt := range tts {
		res, err := index.Search(tt.Search)
		if err != nil {
			t.Errorf("%s - search failed with error: %v", name, err)
		} else if !reflect.DeepEqual(tt.Expected.IDs, res.IDs) {
			t.Errorf("%s - got wrong ids: expected %v got %v", name, tt.Expected, res)
		} else if !reflect.DeepEqual(tt.Expected.Pagination, res.Pagination) {
			t.Errorf("%s - got wrong pagination: expected %v got %v", name, tt.Expected, res)
		}
	}
}
