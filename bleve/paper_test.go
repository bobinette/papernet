package bleve

import (
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/blevesearch/bleve"
)

func createIndex(t *testing.T) (*PaperSearch, func()) {
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

	return &PaperSearch{index: index}, func() {
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

	titles := []string{
		"Title 1",
		"Pizza yolo",
		"titi et rominet",
		"pizza",
	}
	for i, title := range titles {
		data := map[string]interface{}{
			"title": title,
		}
		if err := index.index.Index(strconv.Itoa(i), data); err != nil {
			t.Fatal("error indexing", title, err)
		}
	}

	expected := []int{0, 2}
	ids, err := index.Search("ti")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("got wrong ids: expected %v got %v", expected, ids)
	}
}
