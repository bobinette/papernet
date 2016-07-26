package paper

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

func TestRepository_GetNone(t *testing.T) {
	repo, f := testRepository(t)
	defer f()

	ps, err := repo.Get(1)
	if err != nil {
		t.Fatal("non existing id should not return an error, got", err)
	} else if len(ps) != 0 {
		t.Fatal("got unexpected papers: expected 0 got", len(ps))
	}
}

func TestRepository_Insert(t *testing.T) {
	repo, f := testRepository(t)
	defer f()

	p := Paper{Title: "Test"}
	if err := repo.Insert(&p); err != nil {
		t.Fatal("error inserting:", err)
	}

	ps, err := repo.Get(p.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(ps) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected %d got %d", 1, len(ps))
	} else if g := ps[0]; !reflect.DeepEqual(*g, p) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", p, *g)
	}
}

func TestReposistory_Update(t *testing.T) {
	repo, f := testRepository(t)
	defer f()

	p := Paper{Title: "Test"}
	if err := repo.Insert(&p); err != nil {
		t.Fatal("error inserting:", err)
	}

	p.Title = "Updated"
	if err := repo.Update(&p); err != nil {
		t.Fatal("error inserting:", err)
	}

	ps, err := repo.Get(p.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(ps) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected %d got %d", 1, len(ps))
	} else if g := ps[0]; !reflect.DeepEqual(*g, p) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", p, *g)
	}
}

func testRepository(t *testing.T) (*Repository, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	store, err := bolt.Open(tmpFile.Name(), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		defer os.Remove(tmpFile.Name())
		t.Fatalf("could not open bolt on file %s: %v", tmpFile.Name(), err)
	}

	repo, err := NewRepository(store)
	if err != nil {
		defer os.Remove(tmpFile.Name())
		t.Fatal("could not create repository", err)
	}
	return repo, func() { os.Remove(tmpFile.Name()) }
}
