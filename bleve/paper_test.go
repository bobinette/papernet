package bleve

import (
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/blevesearch/bleve"
)

func createIndex(t *testing.T) (*PaperIndex, func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	indexMapping := createMapping()
	index, err := bleve.New(dir, indexMapping)
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
		Q        string
		Expected []int
	}{
		"match all": {
			Q:        "",
			Expected: []int{0, 1, 2, 3, 4, 5, 6},
		},
		"one word": {
			Q:        "pizza",
			Expected: []int{1, 3, 6},
		},
		"partial word": {
			Q:        "ti",
			Expected: []int{0, 2},
		},
		"two words": {
			Q:        "pizza yolo",
			Expected: []int{1, 6},
		},
		"long words": {
			Q:        "reinforcement learning",
			Expected: []int{4},
		},
		"long words spelling": {
			Q:        "mysuperlnog",
			Expected: []int{},
		},
		"trailing space": {
			Q:        "titi ",
			Expected: []int{2},
		},
		"by tags": {
			Q:        "tech",
			Expected: []int{2, 3},
		},
		"ro": {
			Q:        "pi yo ro",
			Expected: []int{6},
		},
		"with uppercase letters": {
			Q:        "Learning",
			Expected: []int{4, 5},
		},
	}

	for name, tt := range tts {
		ids, err := index.Search(tt.Q)
		if err != nil {
			t.Errorf("%s - search failed with error: %v", name, err)
		} else if !reflect.DeepEqual(tt.Expected, ids) {
			t.Errorf("%s - got wrong ids: expected %v got %v", name, tt.Expected, ids)
		}
	}
}
