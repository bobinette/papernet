package bolt

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet"
)

func createRepository(t *testing.T) (*PaperRepository, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	filename := tmpFile.Name()
	store, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		os.Remove(filename)
		t.Fatalf("could not open bolt on file %s: %v", filename, err)
	}

	err = store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		os.Remove(filename)
		t.Fatal("could not create bucket: ", err)
	}
	repo := &PaperRepository{store: store}

	return repo, func() { os.Remove(filename) }
}

func TestRepository_GetNone(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper, err := repo.Get(1)
	if err != nil {
		t.Fatal("non existing id should not return an error, got", err)
	} else if paper != nil {
		t.Fatal("expected nil, got a non-nil pointer")
	}
}

func TestRepository_Insert(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	retrieved, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if !reflect.DeepEqual(*retrieved, paper) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", paper, *retrieved)
	}
}

func TestReposistory_Update(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	paper.Title = "Updated"
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	retrieved, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if !reflect.DeepEqual(*retrieved, paper) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", paper, *retrieved)
	}
}

func TestRepository_Delete(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	retrieved, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if !reflect.DeepEqual(*retrieved, paper) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", paper, *retrieved)
	}

	err = repo.Delete(paper.ID)
	if err != nil {
		t.Fatal("error deleting", err)
	}
	retrieved, err = repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if retrieved != nil {
		t.Fatalf("incorrect paper retrieved: expected nil got %+v", *retrieved)
	}
}

func TestRepository_List(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	papers := []*papernet.Paper{
		&papernet.Paper{Title: "Test"},
		&papernet.Paper{Title: "Test 2", Summary: "Summary"},
	}
	for _, paper := range papers {
		if err := repo.Upsert(paper); err != nil {
			t.Fatal("error inserting:", err)
		}
	}

	retrieved, err := repo.List()
	if err != nil {
		t.Fatal("error getting:", err)
	} else if !reflect.DeepEqual(retrieved, papers) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", papers, retrieved)
	}
}
