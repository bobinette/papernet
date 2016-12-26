package bleve

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
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

	papers := []struct {
		Title string
		Tags  []string
	}{
		{Title: "Title 1", Tags: []string{"tag"}},
		{Title: "Pizza yolo", Tags: []string{"tag"}},
		{Title: "titi et rominet", Tags: []string{"tag", "tech"}},
		{Title: "pizza", Tags: []string{"tag", "technique"}},
		{Title: "reinforcement learning", Tags: []string{"machine learning"}},
		{Title: "monte carlo", Tags: []string{"machine learning"}},
		{Title: "pizza yolo", Tags: []string{"tag", "robbery"}},
	}
	for i, paper := range papers {
		data := map[string]interface{}{
			"title": paper.Title,
			"tags":  paper.Tags,
		}
		if err := index.index.Index(strconv.Itoa(i), data); err != nil {
			t.Fatal("error indexing", i, err)
		}
	}

	var tts = map[string]struct {
		Search   papernet.PaperSearch
		Expected []int
	}{
		"match all": {
			Search: papernet.PaperSearch{
				Q:   "",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{0, 1, 2, 3, 4, 5, 6},
		},
		"one word": {
			Search: papernet.PaperSearch{
				Q:   "pizza",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{1, 3, 6},
		},
		"partial word": {
			Search: papernet.PaperSearch{
				Q:   "ti",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{0, 2},
		},
		"two words": {
			Search: papernet.PaperSearch{
				Q:   "pizza yolo",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{1, 6},
		},
		"long words": {
			Search: papernet.PaperSearch{
				Q:   "reinforcement learning",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{4},
		},
		"long words spelling": {
			Search: papernet.PaperSearch{
				Q:   "mysuperlnog",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{},
		},
		"trailing space": {
			Search: papernet.PaperSearch{
				Q:   "titi ",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{2},
		},
		"by tags": {
			Search: papernet.PaperSearch{
				Q:   "tech",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{2, 3},
		},
		"ro": {
			Search: papernet.PaperSearch{
				Q:   "pi yo ro",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{6},
		},
		"with uppercase letters": {
			Search: papernet.PaperSearch{
				Q:   "Learning",
				IDs: []int{0, 1, 2, 3, 4, 5, 6},
			},
			Expected: []int{4, 5},
		},
		"by ids": {
			Search: papernet.PaperSearch{
				IDs: []int{1, 3, 17},
			},
			Expected: []int{1, 3},
		},
	}

	for name, tt := range tts {
		ids, err := index.Search(tt.Search)
		if err != nil {
			t.Errorf("%s - search failed with error: %v", name, err)
		} else if !reflect.DeepEqual(tt.Expected, ids) {
			t.Errorf("%s - got wrong ids: expected %v got %v", name, tt.Expected, ids)
		}
	}
}
