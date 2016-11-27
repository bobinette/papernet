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
		{Title: "titi et rominet", Tags: []string{"tag", "test"}},
		{Title: "pizza", Tags: []string{"tag", "tech"}},
		{Title: "mysuperlongwordthat Iwanttomatch", Tags: []string{"tag"}},
		{Title: "mysuperlongwordthat Idonotwanttomatch", Tags: []string{"tag"}},
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
		"one word": {
			Q:        "pizza",
			Expected: []int{1, 3},
		},
		"partial word": {
			Q:        "ti",
			Expected: []int{0, 2},
		},
		"two words": {
			Q:        "pizza yolo",
			Expected: []int{1},
		},
		"long words": {
			Q:        "mysuperlong",
			Expected: []int{4, 5},
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
			Q:        "te",
			Expected: []int{2, 3},
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
